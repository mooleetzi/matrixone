package inner

import (
	"bytes"
	"fmt"
	"matrixbase/pkg/container/batch"
	"matrixbase/pkg/container/vector"
	"matrixbase/pkg/hash"
	"matrixbase/pkg/intmap/fastmap"
	"matrixbase/pkg/vm/mempool"
	"matrixbase/pkg/vm/process"
	"matrixbase/pkg/vm/register"
)

func init() {
	ZeroBools = make([]bool, UnitLimit)
	OneUint64s = make([]uint64, UnitLimit)
	for i := range OneUint64s {
		OneUint64s[i] = 1
	}
}

func String(arg interface{}, buf *bytes.Buffer) {
	n := arg.(*Argument)
	buf.WriteString(fmt.Sprintf("%s ⨝ %s", n.R, n.S))
}

func Prepare(proc *process.Process, arg interface{}) error {
	n := arg.(*Argument)
	n.Ctr = Container{
		builded: false,
		diffs:   make([]bool, UnitLimit),
		matchs:  make([]int64, UnitLimit),
		hashs:   make([]uint64, UnitLimit),
		sels:    make([][]int64, UnitLimit),
		groups:  make(map[uint64][]*hash.SetGroup),
		slots:   fastmap.Pool.Get().(*fastmap.Map),
	}
	return nil
}

func Call(proc *process.Process, arg interface{}) (bool, error) {
	n := arg.(*Argument)
	ctr := &n.Ctr
	if !ctr.builded {
		if err := ctr.build(n.Attrs, proc); err != nil {
			return true, err
		}
		ctr.builded = true
	}
	return ctr.probe(n.R, n.S, n.Attrs, proc)
}

// R ⨝ S - S is the smaller relation
func (ctr *Container) build(attrs []string, proc *process.Process) error {
	var err error

	reg := proc.Reg.Ws[1]
	for {
		v := <-reg.Ch
		if v == nil {
			reg.Wg.Done()
			break
		}
		bat := v.(*batch.Batch)
		if bat.Attrs == nil {
			reg.Wg.Done()
			continue
		}
		bat.Reorder(attrs)
		if err = bat.Prefetch(attrs, bat.Vecs, proc); err != nil {
			ctr.clean(bat, proc)
			reg.Wg.Done()
			return err
		}
		ctr.bats = append(ctr.bats, bat)
		if len(bat.Sels) == 0 {
			if err = ctr.buildBatch(bat.Vecs[:len(attrs)], proc); err != nil {
				ctr.clean(bat, proc)
				reg.Wg.Done()
				return err
			}
		} else {
			if err = ctr.buildBatchSels(bat.Sels, bat.Vecs[:len(attrs)], proc); err != nil {
				ctr.clean(bat, proc)
				reg.Wg.Done()
				return err
			}
		}
		reg.Wg.Done()
	}
	return nil
}

func (ctr *Container) probe(rName, sName string, attrs []string, proc *process.Process) (bool, error) {
	for {
		reg := proc.Reg.Ws[0]
		v := <-reg.Ch
		if v == nil {
			reg.Wg.Done()
			proc.Reg.Ax = nil
			ctr.clean(nil, proc)
			return true, nil
		}
		bat := v.(*batch.Batch)
		if bat.Attrs == nil {
			reg.Wg.Done()
			continue
		}
		if len(ctr.groups) == 0 {
			reg.Ch = nil
			reg.Wg.Done()
			proc.Reg.Ax = nil
			ctr.clean(bat, proc)
			return true, nil
		}
		bat.Reorder(attrs)
		if len(ctr.attrs) == 0 {
			ctr.attrs = make([]string, 0, len(bat.Attrs)+len(ctr.bats[0].Attrs))
			for _, attr := range bat.Attrs {
				ctr.attrs = append(ctr.attrs, rName+"."+attr)
			}
			for _, attr := range ctr.bats[0].Attrs {
				ctr.attrs = append(ctr.attrs, sName+"."+attr)
			}
		}
		ctr.probeState.bat = batch.New(true, ctr.attrs)
		{
			i := 0
			for _, vec := range bat.Vecs {
				ctr.probeState.bat.Vecs[i] = vector.New(vec.Typ)
				i++
			}
			for _, vec := range ctr.bats[0].Vecs {
				ctr.probeState.bat.Vecs[i] = vector.New(vec.Typ)
				i++
			}
		}
		if len(bat.Sels) == 0 {
			if err := ctr.probeBatch(bat, bat.Vecs[:len(attrs)], proc); err != nil {
				reg.Wg.Done()
				ctr.clean(bat, proc)
				return true, err
			}
		} else {
			if err := ctr.probeBatchSels(bat.Sels, bat, bat.Vecs[:len(attrs)], proc); err != nil {
				reg.Wg.Done()
				ctr.clean(bat, proc)
				return true, err
			}
		}
		if ctr.probeState.bat.Vecs[0] == nil {
			reg.Wg.Done()
			bat.Clean(proc)
			continue
		}
		reg.Wg.Done()
		bat.Clean(proc)
		proc.Reg.Ax = ctr.probeState.bat
		ctr.probeState.bat = nil
		return false, nil
	}
}

func (ctr *Container) buildBatch(vecs []*vector.Vector, proc *process.Process) error {
	for i, j := 0, vecs[0].Length(); i < j; i += UnitLimit {
		length := j - i
		if length > UnitLimit {
			length = UnitLimit
		}
		if err := ctr.buildUnit(i, length, nil, vecs, proc); err != nil {
			return err
		}
	}
	return nil
}

func (ctr *Container) buildBatchSels(sels []int64, vecs []*vector.Vector, proc *process.Process) error {
	for i, j := 0, len(sels); i < j; i += UnitLimit {
		length := j - i
		if length > UnitLimit {
			length = UnitLimit
		}
		if err := ctr.buildUnit(0, length, sels[i:i+length], vecs, proc); err != nil {
			return err
		}
	}
	return nil
}

func (ctr *Container) buildUnit(start, count int, sels []int64,
	vecs []*vector.Vector, proc *process.Process) error {
	var err error

	{
		copy(ctr.hashs[:count], OneUint64s[:count])
		if len(sels) == 0 {
			ctr.fillHash(start, count, vecs)
		} else {
			ctr.fillHashSels(count, sels, vecs)
		}
	}
	copy(ctr.diffs[:count], ZeroBools[:count])
	for i, hs := range ctr.slots.Ks {
		for j, h := range hs {
			remaining := ctr.sels[ctr.slots.Vs[i][j]]
			if gs, ok := ctr.groups[h]; ok {
				for _, g := range gs {
					if remaining, err = g.Fill(remaining, ctr.matchs, vecs, ctr.bats, ctr.diffs, proc); err != nil {
						return err
					}
					copy(ctr.diffs[:len(remaining)], ZeroBools[:len(remaining)])
				}
			} else {
				ctr.groups[h] = make([]*hash.SetGroup, 0, 8)
			}
			for len(remaining) > 0 {
				g := hash.NewSetGroup(int64(len(ctr.bats)-1), int64(remaining[0]))
				ctr.groups[h] = append(ctr.groups[h], g)
				if remaining, err = g.Fill(remaining[1:], ctr.matchs, vecs, ctr.bats, ctr.diffs, proc); err != nil {
					return err
				}
				copy(ctr.diffs[:len(remaining)], ZeroBools[:len(remaining)])
			}
			ctr.sels[ctr.slots.Vs[i][j]] = ctr.sels[ctr.slots.Vs[i][j]][:0]
		}
	}
	ctr.slots.Reset()
	return nil
}

func (ctr *Container) probeBatch(bat *batch.Batch, vecs []*vector.Vector, proc *process.Process) error {
	for i, j := 0, vecs[0].Length(); i < j; i += UnitLimit {
		length := j - i
		if length > UnitLimit {
			length = UnitLimit
		}
		if err := ctr.probeUnit(i, length, nil, bat, vecs, proc); err != nil {
			return err
		}
	}
	return nil
}

func (ctr *Container) probeBatchSels(sels []int64, bat *batch.Batch, vecs []*vector.Vector, proc *process.Process) error {
	for i, j := 0, len(sels); i < j; i += UnitLimit {
		length := j - i
		if length > UnitLimit {
			length = UnitLimit
		}
		if err := ctr.probeUnit(0, length, sels[i:i+length], bat, vecs, proc); err != nil {
			return err
		}
	}
	return nil
}

func (ctr *Container) probeUnit(start, count int, sels []int64, bat *batch.Batch,
	vecs []*vector.Vector, proc *process.Process) error {
	var sel int64
	var err error

	{
		copy(ctr.hashs[:count], OneUint64s[:count])
		if len(sels) == 0 {
			ctr.fillHash(start, count, vecs)
		} else {
			ctr.fillHashSels(count, sels, vecs)
		}
	}
	copy(ctr.diffs[:count], ZeroBools[:count])
	for i, hs := range ctr.slots.Ks {
		for j, h := range hs {
			remaining := ctr.sels[ctr.slots.Vs[i][j]]
			if gs, ok := ctr.groups[h]; ok {
				for k := 0; k < len(gs); k++ {
					g := gs[k]
					if sel, remaining, err = g.Probe(remaining, ctr.matchs, vecs, ctr.bats, ctr.diffs, proc); err != nil {
						return err
					}
					if sel >= 0 {
						gs = append(gs[:k], gs[k+1:]...)
						k--
						if len(gs) == 0 {
							delete(ctr.groups, h)
						} else {
							ctr.groups[h] = gs
						}
						{
							for i, vec := range bat.Vecs {
								if ctr.probeState.bat.Vecs[i].Data == nil {
									if err := ctr.probeState.bat.Vecs[i].UnionOne(vec, sel, proc); err != nil {
										return err
									}
									copy(ctr.probeState.bat.Vecs[i].Data[:mempool.CountSize], vec.Data[:mempool.CountSize])
								} else {
									if err := ctr.probeState.bat.Vecs[i].UnionOne(vec, sel, proc); err != nil {
										return err
									}
								}
							}
						}
						{
							k := len(bat.Vecs)
							for _, vec := range ctr.bats[g.Idx].Vecs {
								if ctr.probeState.bat.Vecs[k].Data == nil {
									if err := ctr.probeState.bat.Vecs[k].UnionOne(vec, g.Sel, proc); err != nil {
										return err
									}
									copy(ctr.probeState.bat.Vecs[k].Data[:mempool.CountSize], vec.Data[:mempool.CountSize])
								} else {
									if err := ctr.probeState.bat.Vecs[k].UnionOne(vec, g.Sel, proc); err != nil {
										return err
									}
								}
								k++
							}
						}
					}
					copy(ctr.diffs[:len(remaining)], ZeroBools[:len(remaining)])
				}
			}
			ctr.sels[ctr.slots.Vs[i][j]] = ctr.sels[ctr.slots.Vs[i][j]][:0]
		}
	}
	ctr.slots.Reset()
	return nil
}

func (ctr *Container) fillHash(start, count int, vecs []*vector.Vector) {
	ctr.hashs = ctr.hashs[:count]
	for _, vec := range vecs {
		hash.Rehash(count, ctr.hashs, vec)
	}
	nextslot := 0
	for i, h := range ctr.hashs {
		slot, ok := ctr.slots.Get(h)
		if !ok {
			slot = nextslot
			ctr.slots.Set(h, slot)
			nextslot++
		}
		ctr.sels[slot] = append(ctr.sels[slot], int64(i+start))
	}
}

func (ctr *Container) fillHashSels(count int, sels []int64, vecs []*vector.Vector) {
	var cnt int64

	{
		for i, sel := range sels {
			if i == 0 || sel > cnt {
				cnt = sel
			}
		}
	}
	ctr.hashs = ctr.hashs[:cnt+1]
	for _, vec := range vecs {
		hash.RehashSels(sels[:count], ctr.hashs, vec)
	}
	nextslot := 0
	for i, h := range ctr.hashs {
		slot, ok := ctr.slots.Get(h)
		if !ok {
			slot = nextslot
			ctr.slots.Set(h, slot)
			nextslot++
		}
		ctr.sels[slot] = append(ctr.sels[slot], sels[i])
	}
}

func (ctr *Container) clean(bat *batch.Batch, proc *process.Process) {
	if bat != nil {
		bat.Clean(proc)
	}
	fastmap.Pool.Put(ctr.slots)
	if ctr.probeState.bat != nil {
		ctr.probeState.bat.Clean(proc)
	}
	for _, bat := range ctr.bats {
		bat.Clean(proc)
	}
	register.FreeRegisters(proc)
}
