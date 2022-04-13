package badfs

import (
	"os"

	"github.com/spf13/afero"
)

type BadFile struct {
	afero.File
	goodFile  afero.File
	writeFail *RandomError
	readFail  *RandomError
}

//func NewBadFile(goodFile afero.File) *BadFile {
//	return &BadFile{goodFile: goodFile}
//}

func NewBadFile(goodFile afero.File, readFail *RandomError, writeFail *RandomError) *BadFile {
	return &BadFile{
		goodFile:  goodFile,
		writeFail: writeFail,
		readFail:  readFail,
	}
}

func (b *BadFile) Close() error {
	return b.goodFile.Close()
}

func (b *BadFile) Name() string {
	return b.goodFile.Name()
}

func (b *BadFile) Readdirnames(n int) ([]string, error) {
	return b.goodFile.Readdirnames(n)
}

func (b *BadFile) Readdir(count int) ([]os.FileInfo, error) {
	return b.goodFile.Readdir(count)
}

func (b *BadFile) Stat() (os.FileInfo, error) {
	return b.goodFile.Stat()
}

func (b *BadFile) Sync() error {
	return b.goodFile.Sync()
}

func (b *BadFile) Truncate(size int64) error {
	return b.goodFile.Truncate(size)
}

func (b *BadFile) WriteString(s string) (ret int, err error) {
	return b.goodFile.WriteString(s)
}
