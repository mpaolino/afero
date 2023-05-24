package badfs

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func TestBadFsFileNoErrorWrapper(t *testing.T) {
	const filename = "/fileTest"

	fs := New(afero.NewMemMapFs())
	sourceFile, err := fs.Create(filename)

	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	badFile := NewBadFile(sourceFile, nil, nil, 0)
	if badFile.Name() != filename {
		t.Error("BadFile name does not match the test source file")
	}
}

func TestBadFsDelay(t *testing.T) {
	const filename = "fileTest"
	const latency = 10 * time.Millisecond

	fs := New(afero.NewMemMapFs())
	sourceFile, err := fs.Create(filename)

	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	badFile := NewBadFile(sourceFile, nil, nil, latency)

	startTime := time.Now()
	_, err = badFile.Stat()
	duration := time.Since(startTime)
	if err != nil {
		t.Errorf("File read operation with latency returned error: %s", err)
	}
	if duration < latency {
		t.Error("File read operation didn't took more than the latency duration")
	}

}

func TestBadFsFileClose(t *testing.T) {
	const filename = "/fileTest"
	const writeErrText = "write error"
	fs := New(afero.NewMemMapFs())
	wrappedFile, err := fs.Create(filename)

	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	badFile, ok := wrappedFile.(*BadFile)

	if !ok {
		t.Error("wrappedFile is not a BadFile")
	}

	err = badFile.Close()
	if err != nil {
		t.Error("BadFile could not cleanly close file")

	}

	// Create and Close with write error
	wrappedFile, err = fs.Create(filename)

	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	writeErr := errors.New(writeErrText)

	badFile = wrappedFile.(*BadFile)
	badFile.AddWriteError(writeErr)

	err = badFile.Close()

	if err == nil {
		t.Error("BadFile Close did not return the write error")
	}

	if err.Error() != writeErrText {
		t.Error("Write error text does not match the one configured")
	}

	badFile.DelWriteError()
	err = badFile.Close()

	if err != nil {
		t.Error("BadFile Close returned error when it shouldn't")
	}
}

func TestBadFsFileReaddirnames(t *testing.T) {
	const dirPath = "/test/path"
	const readErrText = "read error"
	fs := New(afero.NewMemMapFs())
	err := fs.MkdirAll(dirPath, 0600)
	if err != nil {
		t.Errorf("Could not create directory: %s", err)
	}
	wrappedFile, err := fs.Open("/")
	if err != nil {
		t.Errorf("Could open directory: %s", err)
	}

	dirNames, err := wrappedFile.Readdirnames(2)
	if err != nil {
		t.Errorf("Readdirnames returned error: %s", err)

	}
	if len(dirNames) != 1 || dirNames[0] != "test" {
		t.Error("Readdirnames returned wrong directory name")
	}
	readErr := errors.New(readErrText)
	badFile, ok := wrappedFile.(*BadFile)

	if !ok {
		t.Error("wrappedFile is not a BadFile")
	}

	badFile.AddReadError(readErr)

	_, err = wrappedFile.Readdirnames(2)
	if err == nil {
		t.Error("BadFile Close did not return the read error")
	}

	if err.Error() != readErrText {
		t.Error("Read error text does not match the one configured")
	}
}

func TestBadFsFileReaddir(t *testing.T) {
	const dirPath = "/test/path"
	const readErrText = "read error"
	fs := New(afero.NewMemMapFs())
	err := fs.MkdirAll(dirPath, 0600)
	if err != nil {
		t.Errorf("Could not create directory: %s", err)
	}
	wrappedFile, err := fs.Open("/")
	if err != nil {
		t.Errorf("Could open directory: %s", err)
	}
	defer wrappedFile.Close()

	fileInfo, err := wrappedFile.Readdir(1)
	if err != nil {
		t.Errorf("Readdirnames returned error: %s", err)

	}
	if len(fileInfo) != 1 && fileInfo[0].Name() != "test" {
		t.Error("Readdir returned wrong directory name")
	}
	readErr := errors.New(readErrText)
	badFile, ok := wrappedFile.(*BadFile)

	if !ok {
		t.Error("wrappedFile is not a BadFile")
	}

	badFile.AddReadError(readErr)

	_, err = wrappedFile.Readdir(2)
	if err == nil {
		t.Error("BadFile Close did not return the read error")
	}

	if err.Error() != readErrText {
		t.Error("Read error text does not match the one configured")
	}
}

func TestBadFsFileStat(t *testing.T) {
	const filePath = "/fileTest"
	const readErrText = "read error"
	fs := New(afero.NewMemMapFs())
	wrappedFile, err := fs.Create(filePath)

	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	badFile, ok := wrappedFile.(*BadFile)

	if !ok {
		t.Error("wrappedFile is not a BadFile")
	}

	fileInfo, err := badFile.Stat()
	if err != nil {
		t.Error("BadFile returned error on Stat")

	}
	if fileInfo.Name() != "fileTest" {
		t.Error("BadFile fileInfo has wrong file name")
	}

	readErr := errors.New(readErrText)

	badFile = wrappedFile.(*BadFile)
	badFile.AddReadError(readErr)

	_, err = badFile.Stat()

	if err == nil {
		t.Error("BadFile Stat did not return the read error")
	}

	if err.Error() != readErrText {
		t.Error("Read error text does not match the one configured")
	}

	badFile.DelReadError()
	fileInfo, err = badFile.Stat()

	if err != nil {
		t.Error("BadFile Stat returned error when it shouldn't")
	}

	if fileInfo.Name() != "fileTest" {
		t.Error("BadFile fileInfo has wrong file name")
	}
}

func TestBadFsFileSync(t *testing.T) {
	const filePath = "/fileTest"
	const writeErrText = "write error"
	fs := New(afero.NewMemMapFs())
	wrappedFile, err := fs.Create(filePath)

	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	badFile, ok := wrappedFile.(*BadFile)

	if !ok {
		t.Error("wrappedFile is not a BadFile")
	}

	err = badFile.Sync()
	if err != nil {
		t.Error("BadFile returned error on Sync")

	}

	writeErr := errors.New(writeErrText)

	badFile = wrappedFile.(*BadFile)
	badFile.AddWriteError(writeErr)

	err = badFile.Sync()

	if err == nil {
		t.Error("BadFile Sync did not return the read error")
	}

	if err.Error() != writeErrText {
		t.Error("Read error text does not match the one configured")
	}

	badFile.DelWriteError()
	err = badFile.Sync()

	if err != nil {
		t.Error("BadFile Stat returned error when it shouldn't")
	}
}

func checkSize(t *testing.T, f afero.File, size int64) {
	dir, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat %q (looking for size %d): %s", f.Name(), size, err)
	}
	if dir.Size() != size {
		t.Errorf("Stat %q: size %d want %d", f.Name(), dir.Size(), size)
	}
}

func TestBadFsFileTruncate(t *testing.T) {
	const writeErrText = "write error"
	const filePath = "/fileTest"
	fs := New(afero.NewMemMapFs())
	wrappedFile, err := fs.Create(filePath)

	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	badFile, ok := wrappedFile.(*BadFile)

	if !ok {
		t.Error("wrappedFile is not a BadFile")
	}
	checkSize(t, badFile, 0)
	badFile.Write([]byte("hello, world\n"))
	checkSize(t, badFile, 13)
	badFile.Truncate(10)
	checkSize(t, badFile, 10)

	writeErr := errors.New(writeErrText)

	badFile = wrappedFile.(*BadFile)
	badFile.AddWriteError(writeErr)

	err = badFile.Truncate(5)

	if err == nil {
		t.Error("BadFile Truncate did not return the read error")
	}

	if err.Error() != writeErrText {
		t.Error("Read error text does not match the one configured")
	}

	badFile.DelWriteError()
	err = badFile.Truncate(5)

	if err != nil {
		t.Error("BadFile Truncate returned error when it shouldn't")
	}
}

func TestBadFsFileWrite(t *testing.T) {
	const filePath = "/fileTest"
	fs := New(afero.NewMemMapFs())
	f, err := fs.Create(filePath)
	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	assert := func(expected bool, v ...interface{}) {
		if !expected {
			t.Helper()
			t.Fatal(v...)
		}
	}
	badFile, ok := f.(*BadFile)
	assert(ok == true, ok)

	data4 := []byte{0, 1, 2, 3}
	n, err := badFile.Write(data4)
	assert(err == nil, err)
	assert(n == len(data4), n)
	wErr := errors.New("write error")
	badFile.AddWriteError(wErr)
	n, err = badFile.Write(data4)
	assert(err == wErr, err)
	assert(n == -1, n)
	badFile.DelWriteError()
}

func TestBadFsFileSeek(t *testing.T) {
	const filePath = "/fileTest"
	fs := New(afero.NewMemMapFs())
	f, err := fs.Create(filePath)
	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	assert := func(expected bool, v ...interface{}) {
		if !expected {
			t.Helper()
			t.Fatal(v...)
		}
	}
	badFile, ok := f.(*BadFile)
	assert(ok == true, ok)

	data4 := []byte{0, 1, 2, 3}
	n, err := badFile.Write(data4)
	assert(err == nil, err)
	assert(n == len(data4), n)
	rErr := errors.New("read error")
	badFile.AddReadError(rErr)
	var off int64
	off += int64(n)
	n64, err := badFile.Seek(-off, io.SeekCurrent)
	assert(err == rErr, err)
	assert(n64 == -1, n64)
	badFile.DelReadError()
	n64, err = badFile.Seek(-off, io.SeekCurrent)
	assert(err == nil, err)
	assert(n64 == 0, n64)
}
