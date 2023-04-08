package badfs

import (
	"errors"
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
