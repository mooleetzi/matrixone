// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dispatch

import (
	"bytes"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

func String(arg any, buf *bytes.Buffer) {
	buf.WriteString("dispatch")
}

func Prepare(proc *process.Process, arg any) error {
	ap := arg.(*Argument)
	ap.ctr = new(container)

	switch ap.FuncId {
	case SendToAllFunc:
		if len(ap.RemoteRegs) == 0 {
			return moerr.NewInternalError(proc.Ctx, "SendToAllFunc should include RemoteRegs")
		}
		ap.prepared = false
		ap.ctr.remoteReceivers = make([]*WrapperClientSession, 0, len(ap.RemoteRegs))
		ap.ctr.sendFunc = sendToAllFunc
		for _, rr := range ap.RemoteRegs {
			colexec.Srv.PutNotifyChIntoUuidMap(rr.Uuid, proc.DispatchNotifyCh)
		}

	case SendToAllLocalFunc:
		if len(ap.RemoteRegs) != 0 {
			return moerr.NewInternalError(proc.Ctx, "SendToAllLocalFunc should not send to remote")
		}
		ap.prepared = true
		ap.ctr.remoteReceivers = nil
		ap.ctr.sendFunc = sendToAllLocalFunc
	case SendToAnyLocalFunc:
		if len(ap.RemoteRegs) != 0 {
			return moerr.NewInternalError(proc.Ctx, "SendToAnyLocalFunc should not send to remote")
		}
		ap.prepared = true
		ap.ctr.remoteReceivers = nil
		ap.ctr.sendFunc = sendToAnyLocalFunc
	default:
		return moerr.NewInternalError(proc.Ctx, "wrong sendFunc id for dispatch")
	}

	return nil
}

func Call(idx int, proc *process.Process, arg any, isFirst bool, isLast bool) (bool, error) {
	ap := arg.(*Argument)

	// waiting all remote receive prepared
	// put it in Call() for better parallel

	bat := proc.InputBatch()
	if bat == nil {
		return true, nil
	}

	if bat.Length() == 0 {
		return false, nil
	}

	/*
		for i, vec := range bat.Vecs {
			if vec.IsOriginal() {
				cloneVec, err := vec.Dup(proc.Mp())
				if err != nil {
					bat.Clean(proc.Mp())
					return false, err
				}
				bat.Vecs[i] = cloneVec
			}
		}
	*/

	if err := ap.ctr.sendFunc(bat, ap, proc); err != nil {
		return false, err
	}
	if len(ap.LocalRegs) == 0 {
		return true, nil
	}
	proc.SetInputBatch(nil)
	return false, nil
}

func (arg *Argument) waitRemoteRegsReady(proc *process.Process) {
	cnt := len(arg.RemoteRegs)
	for cnt > 0 {
		csinfo := <-proc.DispatchNotifyCh
		arg.ctr.remoteReceivers = append(arg.ctr.remoteReceivers, &WrapperClientSession{
			msgId:  csinfo.MsgId,
			cs:     csinfo.Cs,
			uuid:   csinfo.Uid,
			doneCh: csinfo.DoneCh,
		})
		cnt--
	}
	arg.prepared = true
}
