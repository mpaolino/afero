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
	sourceFile, err := fs.Create(filename)

	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	badFile := NewBadFile(sourceFile, nil, nil, 0)
	if badFile == nil {

		t.Error("BadFile name does not match the test source file")
	}

	err = badFile.Close()
	if err != nil {
		t.Error("BadFile could not cleanly close file")

	}

	// Create and Close with write error
	sourceFile, err = fs.Create(filename)

	if err != nil {
		t.Errorf("Could not create file: %s", err)
	}

	writeErr := errors.New(writeErrText)

	badFile = NewBadFile(sourceFile, nil, writeErr, 0)
	badFile.AddWriteError(writeErr)

	err = badFile.Close()

	if err == nil {
		t.Error("BadFile Close did not return the added write error")
	}

	if err.Error() != writeErrText {
		t.Error("Write error text does not match the one configured")
	}

	badFile.DelWriteError()
	err = badFile.Close()

	if err != nil {
		t.Error("BadFile Close did returned error when there should be none")
	}
}
