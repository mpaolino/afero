package badfs

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"os"
	"sync"
	"syscall"
	"time"
)

type errorMap map[string]error
type latencyMap map[string]time.Duration

type BadFs struct {
	source      afero.Fs
	writeErrors errorMap
	readErrors  errorMap
	latencies   latencyMap
	mu          sync.RWMutex
}

func New(source afero.Fs) *BadFs {
	return &BadFs{
		source:      source,
		writeErrors: errorMap{},
		readErrors:  errorMap{},
		latencies:   latencyMap{},
		mu:          sync.RWMutex{},
	}
}

func normalizePath(path string) string {
	return filepath.Clean(path)
}

func (r *BadFs) AddWriteError(name string, err error) {
	name = normalizePath(name)
	r.mu.Lock()
	r.writeErrors[name] = err
	r.mu.Unlock()
}

func (r *BadFs) DelWriteError(name string) {
	name = normalizePath(name)
	r.mu.Lock()
	delete(r.writeErrors, name)
	r.mu.Unlock()
}

func (r *BadFs) AddReadError(name string, err error) {
	name = normalizePath(name)
	r.mu.Lock()
	r.readErrors[name] = err
	r.mu.Unlock()
}

func (r *BadFs) DelReadError(name string) {
	name = normalizePath(name)
	r.mu.Lock()
	delete(r.readErrors, name)
	r.mu.Unlock()
}

func (r *BadFs) AddLatency(name string, latency time.Duration) error {
	name = normalizePath(name)

	if latency <= 0 {
		return fmt.Errorf("latency for I/O operations should be positive time durations")
	}
	r.mu.Lock()
	r.latencies[name] = latency
	r.mu.Unlock()
	return nil
}

func (r *BadFs) DelLatency(name string) {
	name = normalizePath(name)
	r.mu.Lock()
	delete(r.latencies, name)
	r.mu.Unlock()
}

func (r *BadFs) GetLatency(name string) (time.Duration, error) {
	name = normalizePath(name)
	r.mu.RLock()
	defer r.mu.RUnlock()
	if latency, hasLatency := r.latencies[name]; hasLatency {
		return latency, nil

	}
	return 0, fmt.Errorf("no latency registered for '%s'", name)
}

func (r *BadFs) getLatencies() latencyMap {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.latencies
}

func (r *BadFs) getError(errMap errorMap, name string) (error, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if opErr, hasError := errMap[name]; hasError && opErr != nil {
		return opErr, nil

	}
	return nil, fmt.Errorf("no error registered for '%s'", name)
}

func (r *BadFs) getWriteErrors() errorMap {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.writeErrors
}

func (r *BadFs) GetWriteError(name string) (error, error) {
	name = normalizePath(name)
	return r.getError(r.writeErrors, name)
}

func (r *BadFs) GetReadErrors() errorMap {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.readErrors
}

func (r *BadFs) GetReadError(name string) (error, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	name = normalizePath(name)
	return r.getError(r.readErrors, name)
}

func (r *BadFs) delay(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if latency, hasLatency := r.latencies[name]; hasLatency {
		time.Sleep(latency)
	}
}

func (r *BadFs) checkError(errMap errorMap, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if opError, hasError := errMap[name]; hasError && opError != nil {
		return opError
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
	name = normalizePath(name)
	r.delay(name)

	if err := r.checkWriteError(name); err != nil {
		return err
	}
	return nil
}

func (r *BadFs) readOperation(name string) error {
	name = normalizePath(name)
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
	n = normalizePath(n)
	if err := r.writeOperation(n); err != nil {
		return err
	}

	return r.source.Chtimes(n, a, m)
}

func (r *BadFs) Chmod(n string, m os.FileMode) error {
	n = normalizePath(n)
	if err := r.writeOperation(n); err != nil {
		return err
	}
	return r.source.Chmod(n, m)
}

func (r *BadFs) Chown(n string, uid, gid int) error {
	n = normalizePath(n)

	if err := r.writeOperation(n); err != nil {
		return err
	}
	return r.source.Chown(n, uid, gid)
}

func (r *BadFs) Name() string {
	return "BadFsWrapper"
}

func (r *BadFs) Stat(name string) (os.FileInfo, error) {
	name = normalizePath(name)
	if err := r.readOperation(name); err != nil {
		return nil, err
	}
	return r.source.Stat(name)
}

func (r *BadFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	name = normalizePath(name)
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

func (r *BadFs) SymlinkIfPossible(name, linkName string) error {
	name = normalizePath(name)
	linkName = normalizePath(linkName)

	slayer, symlinkOk := r.source.(afero.Linker)

	if err := r.writeOperation(name); err != nil {
		return err
	}

	if !symlinkOk {
		return &os.LinkError{Op: "symlink", Old: name, New: linkName, Err: afero.ErrNoSymlink}
	}

	if err := slayer.SymlinkIfPossible(name, linkName); err != nil {
		return err
	}

	//Symlink successfull, let's add the same errors for the new file
	r.copyErrors(name, linkName)
	return nil
}

func (r *BadFs) ReadlinkIfPossible(name string) (string, error) {
	name = normalizePath(name)

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
	o = normalizePath(o)
	n = normalizePath(n)

	if err := r.writeOperation(o); err != nil {
		return err
	}
	return r.source.Rename(o, n)
}

func (r *BadFs) RemoveAll(path string) error {
	path = normalizePath(path)

	for p := range r.getLatencies() {
		if p == path || strings.HasPrefix(p, path+afero.FilePathSeparator) {
			r.delay(p)
		}
	}

	for p, err := range r.getWriteErrors() {
		if p == path || strings.HasPrefix(p, path+afero.FilePathSeparator) {
			return err
		}
	}
	return r.source.RemoveAll(path)
}

func (r *BadFs) Remove(n string) error {
	n = normalizePath(n)

	if err := r.writeOperation(n); err != nil {
		return err
	}
	return r.source.Remove(n)
}

func (r *BadFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	name = normalizePath(name)

	// Determine whether it's a write operation
	isWrite := flag&(os.O_WRONLY|syscall.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0

	// Call the appropriate operation function
	var opErr error
	if isWrite {
		opErr = r.writeOperation(name)
	} else {
		opErr = r.readOperation(name)
	}

	// Return the error if there is one
	if opErr != nil {
		return nil, opErr
	}

	sourceFile, err := r.source.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	return NewBadFile(sourceFile, r.readErrors[name], r.writeErrors[name], r.latencies[name]), nil
}

func (r *BadFs) Open(name string) (afero.File, error) {
	name = normalizePath(name)
	if err := r.readOperation(name); err != nil {
		return nil, err
	}

	sourceFile, err := r.source.Open(name)

	if err != nil {
		return nil, err
	}
	return NewBadFile(sourceFile, r.readErrors[name], r.writeErrors[name], r.latencies[name]), nil
}

func (r *BadFs) Mkdir(n string, p os.FileMode) error {
	n = normalizePath(n)

	if err := r.writeOperation(n); err != nil {
		return err
	}
	return r.source.Mkdir(n, p)
}

func (r *BadFs) MkdirAll(path string, perm os.FileMode) error {
	path = normalizePath(path)

	for p := range r.getLatencies() {
		if p == path || strings.HasPrefix(p, path+afero.FilePathSeparator) {
			r.delay(p)
		}
	}

	for p, err := range r.getWriteErrors() {
		if p == path || strings.HasPrefix(p, path+afero.FilePathSeparator) {
			return err
		}
	}

	return r.source.MkdirAll(path, perm)

}

func (r *BadFs) Create(name string) (afero.File, error) {
	if err := r.writeOperation(name); err != nil {
		return nil, err
	}

	sourceFile, err := r.source.Create(name)

	if err != nil {
		return nil, err
	}
	return NewBadFile(sourceFile, r.readErrors[name], r.writeErrors[name], r.latencies[name]), nil
}
