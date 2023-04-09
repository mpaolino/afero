package badfs

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/afero"
)

type BadFile struct {
	afero.File
	sourceFile afero.File
	writeError error
	readError  error
	latency    time.Duration
	mu         sync.RWMutex
}

//func NewBadFile(goodFile afero.File) *BadFile {
//	return &BadFile{goodFile: goodFile}
//}

func NewBadFile(goodFile afero.File, readError error, writeError error, latency time.Duration) *BadFile {
	return &BadFile{
		sourceFile: goodFile,
		writeError: writeError,
		readError:  readError,
		latency:    latency,
		mu:         sync.RWMutex{},
	}
}

func (b *BadFile) AddLatency(latency time.Duration) error {
	if latency < 0 {
		return fmt.Errorf("latency for I/O operations should be positive time durations")
	}
	b.mu.Lock()
	b.latency = latency
	b.mu.Unlock()
	return nil
}

func (b *BadFile) GetLatency() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.latency
}

func (b *BadFile) delay() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.latency > 0 {
		time.Sleep(b.latency)
	}
}

func (b *BadFile) AddWriteError(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.writeError = err
}

func (b *BadFile) DelWriteError() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.writeError = nil
}

func (b *BadFile) AddReadError(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.readError = err
}

func (b *BadFile) DelReadError() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.readError = nil
}

func (b *BadFile) getWriteError() error {
	return b.writeError
}

func (b *BadFile) getReadError() error {
	return b.readError
}

func (b *BadFile) Close() error {
	b.delay()
	if err := b.getWriteError(); err != nil {
		return err
	}
	return b.sourceFile.Close()
}

func (b *BadFile) Name() string {
	b.delay()
	return b.sourceFile.Name()
}

func (b *BadFile) Readdirnames(n int) ([]string, error) {
	b.delay()
	if err := b.getReadError(); err != nil {
		return nil, err
	}
	return b.sourceFile.Readdirnames(n)
}

func (b *BadFile) Readdir(count int) ([]os.FileInfo, error) {
	b.delay()
	if err := b.getReadError(); err != nil {
		return nil, err
	}
	return b.sourceFile.Readdir(count)
}

func (b *BadFile) Stat() (os.FileInfo, error) {
	b.delay()
	if err := b.getReadError(); err != nil {
		return nil, err
	}
	return b.sourceFile.Stat()
}

func (b *BadFile) Sync() error {
	b.delay()
	if err := b.getWriteError(); err != nil {
		return err
	}
	return b.sourceFile.Sync()
}

func (b *BadFile) Truncate(size int64) error {
	b.delay()
	if err := b.getWriteError(); err != nil {
		return err
	}
	return b.sourceFile.Truncate(size)
}

func (b *BadFile) Write(wb []byte) (n int, err error) {
	b.delay()
	if err := b.getWriteError(); err != nil {
		return -1, err
	}
	return b.sourceFile.Write(wb)
}

func (b *BadFile) WriteAt(wb []byte, off int64) (n int, err error) {
	b.delay()
	if err := b.getWriteError(); err != nil {
		return -1, err
	}
	return b.sourceFile.WriteAt(wb, off)
}

func (b *BadFile) WriteString(s string) (ret int, err error) {
	b.delay()
	if err := b.getWriteError(); err != nil {
		return -1, err
	}
	return b.sourceFile.WriteString(s)
}

func (b *BadFile) Read(rb []byte) (n int, err error) {
	b.delay()
	if err := b.getReadError(); err != nil {
		return -1, err
	}
	return b.sourceFile.Read(rb)
}

func (b *BadFile) ReadAt(rb []byte, off int64) (n int, err error) {
	b.delay()
	if err := b.getReadError(); err != nil {
		return -1, err
	}
	return b.sourceFile.ReadAt(rb, off)
}

func (b *BadFile) Seek(offset int64, whence int) (int64, error) {
	b.delay()
	if err := b.getReadError(); err != nil {
		return -1, err
	}
	return b.sourceFile.Seek(offset, whence)
}
