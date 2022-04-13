package badfs

import (
	"fmt"

	"github.com/spf13/afero"

	"os"
	"sync"
	"syscall"
	"time"
)

type BadFs struct {
	source      afero.Fs
	writeErrors map[string]*RandomError
	readErrors  map[string]*RandomError
	latencies   map[string]time.Duration
	mu          sync.RWMutex
}

func New(source afero.Fs) *BadFs {
	return &BadFs{
		source:      source,
		writeErrors: map[string]*RandomError{},
		readErrors:  map[string]*RandomError{},
		latencies:   map[string]time.Duration{},
		mu:          sync.RWMutex{},
	}
}

func (r *BadFs) AddRandomWriteError(name string, err error, probability float64) {
	r.mu.Lock()
	r.writeErrors[name] = NewRandomError(err, probability)
	r.mu.Unlock()
}

func (r *BadFs) AddWriteError(name string, err error) {
	r.AddRandomWriteError(name, err, 1)
}

func (r *BadFs) DelWriteError(name string) {
	r.mu.Lock()
	delete(r.writeErrors, name)
	r.mu.Unlock()
}

func (r *BadFs) AddRandomReadError(name string, err error, probability float64) {
	r.mu.Lock()
	r.readErrors[name] = NewRandomError(err, probability)
	r.mu.Unlock()
}

func (r *BadFs) AddReadError(name string, err error) {
	r.AddRandomReadError(name, err, 1)
}

func (r *BadFs) DelReadError(name string) {
	r.mu.Lock()
	delete(r.readErrors, name)
	r.mu.Unlock()
}

func (r *BadFs) AddLatency(name string, latency time.Duration) error {
	if latency <= 0 {
		return fmt.Errorf("latency for I/O operations should be positive time durations")
	}

	r.latencies[name] = latency
	return nil
}

func (r *BadFs) DelLatency(name string) {
	delete(r.latencies, name)
}

func (r *BadFs) GetLatency(name string) (time.Duration, error) {
	if latency, hasLatency := r.latencies[name]; hasLatency {
		return latency, nil

	}
	return 0, fmt.Errorf("no latency registered for '%s'", name)
}

func (r *BadFs) getError(errMap map[string]*RandomError, name string) (error, error) {
	if rErr, hasError := errMap[name]; hasError {
		return rErr.getError(), nil

	}
	return nil, fmt.Errorf("no error registered for '%s'", name)
}

func (r *BadFs) GetWriteError(name string) (error, error) {
	return r.getError(r.writeErrors, name)
}

func (r *BadFs) GetReadError(name string) (error, error) {
	return r.getError(r.readErrors, name)
}

func (r *BadFs) delay(name string) {
	if latency, hasLatency := r.latencies[name]; hasLatency {
		time.Sleep(latency)
	}
}

func checkError(errors map[string]*RandomError, name string) error {
	if randomError, hasError := errors[name]; hasError {
		return randomError.getError()
	}
	return nil
}

func (r *BadFs) checkWriteError(name string) error {
	return checkError(r.writeErrors, name)
}

func (r *BadFs) checkReadError(name string) error {
	return checkError(r.readErrors, name)
}

func (r *BadFs) writeOperation(name string) error {
	r.delay(name)

	if err := r.checkWriteError(name); err != nil {
		return err
	}
	return nil
}

func (r *BadFs) readOperation(name string) error {
	r.delay(name)

	if err := r.checkReadError(name); err != nil {
		return err
	}
	return nil
}

// afero Fs interface implementation

func (r *BadFs) Chtimes(n string, a, m time.Time) error {
	if err := r.writeOperation(n); err != nil {
		return err
	}

	return r.source.Chtimes(n, a, m)
}

func (r *BadFs) Chmod(n string, m os.FileMode) error {
	if err := r.writeOperation(n); err != nil {
		return err
	}
	return r.source.Chmod(n, m)
}

func (r *BadFs) Chown(n string, uid, gid int) error {
	if err := r.writeOperation(n); err != nil {
		return err
	}
	return r.source.Chown(n, uid, gid)
}

func (r *BadFs) Name() string {
	return "BadFsWrapper"
}

func (r *BadFs) Stat(name string) (os.FileInfo, error) {
	if err := r.readOperation(name); err != nil {
		return nil, err
	}
	return r.source.Stat(name)
}

/*
func (r *BadFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	if lsf, ok := r.source.(Lstater); ok {
		return lsf.LstatIfPossible(name)
	}
	fi, err := r.Stat(name)
	return fi, false, err
}


func (r *BadFs) SymlinkIfPossible(oldname, newname string) error {
	return &os.LinkError{Op: "symlink", Old: oldname, New: newname, Err: ErrNoSymlink}
}

func (r *BadFs) ReadlinkIfPossible(name string) (string, error) {
	if srdr, ok := r.source.(LinkReader); ok {
		return srdr.ReadlinkIfPossible(name)
	}

	return "", &os.PathError{Op: "readlink", Path: name, Err: ErrNoReadlink}
}
*/

func (r *BadFs) Rename(o, n string) error {
	if err := r.writeOperation(n); err != nil {
		return err
	}
	return r.source.Rename(o, n)
}

func (r *BadFs) RemoveAll(p string) error {
	if err := r.writeOperation(p); err != nil {
		return err
	}
	return r.source.RemoveAll(p)
}

func (r *BadFs) Remove(n string) error {
	if err := r.writeOperation(n); err != nil {
		return err
	}
	return r.source.Remove(n)
}

func (r *BadFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_WRONLY|syscall.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		if err := r.writeOperation(name); err != nil {
			return nil, err
		}
		return r.source.OpenFile(name, flag, perm)
	}

	if err := r.readOperation(name); err != nil {
		return nil, err
	}
	return r.source.OpenFile(name, flag, perm)
}

func (r *BadFs) Open(name string) (afero.File, error) {
	if err := r.readOperation(name); err != nil {
		return nil, err
	}

	sourceFile, err := r.source.Open(name)

	if err != nil {
		return nil, err
	}

	return NewBadFile(sourceFile, r.readErrors[name], r.writeErrors[name]), nil
}

func (r *BadFs) Mkdir(n string, p os.FileMode) error {
	if err := r.writeOperation(n); err != nil {
		return err
	}
	return r.source.Mkdir(n, p)
}

func (r *BadFs) MkdirAll(n string, p os.FileMode) error {
	if err := r.writeOperation(n); err != nil {
		return err
	}
	return r.source.MkdirAll(n, p)
}

func (r *BadFs) Create(name string) (afero.File, error) {
	if err := r.writeOperation(name); err != nil {
		return nil, err
	}
	return r.source.Create(name)
}
