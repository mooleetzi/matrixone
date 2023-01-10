// Copyright 2022 Matrix Origin
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

package export

import (
	"bytes"
	"context"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/util/export/table"
	"github.com/matrixorigin/matrixone/pkg/util/export/writer"
	"github.com/matrixorigin/matrixone/pkg/util/trace/impl/motrace"
	"io"
	"path"
	"sync"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/defines"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
)

var _ stringWriter = (*FSWriter)(nil)

type FSWriter struct {
	ctx context.Context         // New args
	fs  fileservice.FileService // New args
	// filepath
	filepath string // see WithFilePath or auto generated by NewFSWriter

	mux sync.Mutex

	fileServiceName string // const

	offset int // see Write, should not have size bigger than 2GB
}

type FSWriterOption func(*FSWriter)

func (f FSWriterOption) Apply(w *FSWriter) {
	f(w)
}

func NewFSWriter(ctx context.Context, fs fileservice.FileService, opts ...FSWriterOption) *FSWriter {
	w := &FSWriter{
		ctx:             ctx,
		fs:              fs,
		fileServiceName: defines.ETLFileServiceName,
	}
	for _, o := range opts {
		o.Apply(w)
	}
	if len(w.filepath) == 0 {
		panic("filepath is Empty")
	}
	return w
}

func WithFilePath(filepath string) FSWriterOption {
	return FSWriterOption(func(w *FSWriter) {
		w.filepath = filepath
	})
}

// Write implement io.Writer, Please execute in series
func (w *FSWriter) Write(p []byte) (n int, err error) {
	w.mux.Lock()
	defer w.mux.Unlock()
	n = len(p)
	mkdirTried := false
mkdirRetry:
	if err = w.fs.Write(w.ctx, fileservice.IOVector{
		// like: etl:store/system/filename.csv
		FilePath: w.fileServiceName + fileservice.ServiceNameSeparator + w.filepath,
		Entries: []fileservice.IOEntry{
			{
				Offset: int64(w.offset),
				Size:   int64(n),
				Data:   p,
			},
		},
	}); err == nil {
		w.offset += n
	} else if moerr.IsMoErrCode(err, moerr.ErrFileAlreadyExists) && !mkdirTried {
		mkdirTried = true
		goto mkdirRetry
	}
	// XXX Why call this?
	// _ = errors.WithContext(w.ctx, err)
	return
}

// WriteString implement io.StringWriter
func (w *FSWriter) WriteString(s string) (n int, err error) {
	var b = String2Bytes(s)
	return w.Write(b)
}

type FSWriterFactory0 func(ctx context.Context, account string, tbl *table.Table, ts time.Time) io.StringWriter
type FSWriterFactory func(ctx context.Context, account string, tbl *table.Table, ts time.Time) table.RowWriter

func GetFSWriterFactory(fs fileservice.FileService, nodeUUID, nodeType, ext string) FSWriterFactory0 {
	var extension = table.GetExtension(ext)
	var cfg = FilePathCfg{NodeUUID: nodeUUID, NodeType: nodeType, Extension: extension}
	return func(ctx context.Context, account string, tbl *table.Table, ts time.Time) io.StringWriter {
		return NewFSWriter(ctx, fs, WithFilePath(cfg.LogsFilePathFactory(account, tbl, ts)))
	}
}

type FilePathCfg struct {
	NodeUUID  string
	NodeType  string
	Extension string
}

func (c *FilePathCfg) LogsFilePathFactory(account string, tbl *table.Table, ts time.Time) string {
	filename := tbl.PathBuilder.NewLogFilename(tbl.Table, c.NodeUUID, c.NodeType, ts, c.Extension)
	dir := tbl.PathBuilder.Build(account, table.MergeLogTypeLogs, ts, tbl.Database, tbl.Table)
	return path.Join(dir, filename)
}

func GetRowWriterFactory4Trace(fs fileservice.FileService, nodeUUID, nodeType string, ext string) (factory motrace.FSWriterFactory) {

	var extension = table.GetExtension(ext)
	var cfg = FilePathCfg{NodeUUID: nodeUUID, NodeType: nodeType, Extension: extension}

	switch extension {
	case table.CsvExtension:
		factory = func(ctx context.Context, account string, tbl *table.Table, ts time.Time) table.RowWriter {
			options := []FSWriterOption{
				WithFilePath(cfg.LogsFilePathFactory(account, tbl, ts)),
			}
			//if name != nil {
			//	options = append(options, WithName(name))
			//}
			return writer.NewCSVWriter(ctx, bytes.NewBuffer(nil), NewFSWriter(ctx, fs, options...))
		}
	case table.TaeExtension:
		mp, err := mpool.NewMPool("etl_fs_writer", 0, mpool.NoFixed)
		if err != nil {
			panic(err)
		}
		factory = func(ctx context.Context, account string, tbl *table.Table, ts time.Time) table.RowWriter {
			filePath := cfg.LogsFilePathFactory(account, tbl, ts)
			return writer.NewTAEWriter(ctx, tbl, mp, filePath, fs)
		}
	}

	return factory
}
