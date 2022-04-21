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
	r.mu.Lock()
	r.latencies[name] = latency
	r.mu.Unlock()
	return nil
}

func (r *BadFs) DelLatency(name string) {
	r.mu.Lock()
	delete(r.latencies, name)
	r.mu.Unlock()
}

func (r *BadFs) GetLatency(name string) (time.Duration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if latency, hasLatency := r.latencies[name]; hasLatency {
		return latency, nil

	}
	return 0, fmt.Errorf("no latency registered for '%s'", name)
}

func (r *BadFs) getError(errMap map[string]*RandomError, name string) (error, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rErr, hasError := errMap[name]; hasError && rErr != nil {
		print(fmt.Errorf("rErr: %s", rErr.err.Error()))
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
	r.mu.Lock()
	defer r.mu.Unlock()
	if latency, hasLatency := r.latencies[name]; hasLatency {
		time.Sleep(latency)
	}
}

func (r *BadFs) checkError(errorMap map[string]*RandomError, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if randomError, hasError := errorMap[name]; hasError && randomError != nil {
		return randomError.getError()
	}
	return nil
}

func (r *BadFs) checkWriteError(name string) error {
	return r.checkError(r.writeErrors, name)
}

func (r *BadFs) checkReadError(name string) error {
	return r.checkError(r.readErrors, name)
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

//
// afero Fs interface implementation, including optional interfaces
//

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

func (r *BadFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {

	lsf, lsf_ok := r.source.(afero.Lstater)

	if err := r.readOperation(name); err != nil {
		return nil, lsf_ok, err
	}

	if lsf_ok {
		return lsf.LstatIfPossible(name)
	}

	fi, err := r.source.Stat(name)
	return fi, lsf_ok, err
}

func (r *BadFs) copyErrors(src, dst string) {
	r.writeErrors[dst] = r.writeErrors[src]
	r.readErrors[dst] = r.readErrors[src]
	r.latencies[dst] = r.latencies[src]
}

func (r *BadFs) SymlinkIfPossible(name, linkname string) error {

	slayer, symlinkOk := r.source.(afero.Linker)

	if err := r.writeOperation(name); err != nil {
		return err
	}

	if !symlinkOk {
		return &os.LinkError{Op: "symlink", Old: name, New: linkname, Err: afero.ErrNoSymlink}
	}

	if err := slayer.SymlinkIfPossible(name, linkname); err != nil {
		return err
	}

	//Symlink successfull, let's add the same errors for the new file that were in the target file
	r.copyErrors(name, linkname)
	return nil
}

func (r *BadFs) ReadlinkIfPossible(name string) (string, error) {
	srdr, rlink_ok := r.source.(afero.LinkReader)

	if err := r.readOperation(name); err != nil {
		return "", err
	}

	if rlink_ok {
		return srdr.ReadlinkIfPossible(name)
	}

	return "", &os.PathError{Op: "readlink", Path: name, Err: afero.ErrNoReadlink}
}

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
