package badfs

import (
	"errors"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func TestBadFsEmptyWrapper(t *testing.T) {
	const filename = "/fileTest"
	fs := New(afero.NewMemMapFs())
	file, err := fs.Create(filename)
	if err != nil {
		t.Error("Could not create file")
	}
	if file.Name() != filename {
		t.Error("Test filename does not match the one returned")
	}

}

func TestBadFsAddWriteError(t *testing.T) {
	fs := New(afero.NewMemMapFs())
	const errorDesc = "write test error"
	fs.AddWriteError("/", errors.New(errorDesc))
	writeErr, err := fs.GetWriteError("/")
	if err != nil {
		t.Error("Could not retrieve created write error")
	}
	if writeErr.Error() != errorDesc {
		t.Error("Retreived error does not match the added error")
	}
}

func TestBadFsAddReadError(t *testing.T) {
	fs := New(afero.NewMemMapFs())
	const errorDesc = "read test error"
	fs.AddReadError("/", errors.New(errorDesc))
	readErr, err := fs.GetReadError("/")
	if err != nil {
		t.Error("Could not retrieve created read error")
	}
	if readErr.Error() != errorDesc {
		t.Error("Retreived error does not match the added error")
	}
}

func TestBadFsGetWriteErrorFailure(t *testing.T) {
	fs := New(afero.NewMemMapFs())
	const errorDesc = "write test error"
	fs.AddWriteError("/", errors.New(errorDesc))
	writeErr, err := fs.GetReadError("/notroot")
	if err == nil {
		t.Error("Error should have being returned")
	}
	if writeErr != nil {
		t.Error("Write error should not have being returned")
	}
}

func TestBadFsGetReadErrorFailure(t *testing.T) {
	fs := New(afero.NewMemMapFs())
	const errorDesc = "read test error"
	fs.AddReadError("/", errors.New(errorDesc))
	readErr, err := fs.GetReadError("/notroot")
	if err == nil {
		t.Error("Error should have being returned")
	}
	if readErr != nil {
		t.Error("Read error should not have being returned")
	}
}

func TestBadFsWriteError(t *testing.T) {
	const filename = "/myTestFile"
	const errorDesc = "write test error"
	fs := New(afero.NewMemMapFs())
	fs.AddWriteError(filename, errors.New(errorDesc))
	file, err := fs.Create(filename)

	if err.Error() != errorDesc {
		t.Errorf("Wrong write error returned: %s", err.Error())
	}

	if file != nil {
		t.Error("File should have not being created in this path")
	}
}

func TestBadFsReadError(t *testing.T) {
	const filename = "/myTestFile"
	const errorDesc = "read test error"
	fs := New(afero.NewMemMapFs())
	fs.AddReadError(filename, errors.New(errorDesc))
	_, err := fs.Create(filename)
	if err != nil {
		t.Error("Could not create file")
	}

	fileInfo, err := fs.Stat(filename)

	if err.Error() != errorDesc {
		t.Errorf("Read error not returned for file")
	}

	if fileInfo != nil {
		t.Error("FileInfo should not have being returned")
	}

}

func TestBadFsLatency(t *testing.T) {
	const filename = "myTestFile"
	const latency = 10 * time.Millisecond
	fs := New(afero.NewMemMapFs())
	_, err := fs.Create(filename)
	if err != nil {
		t.Error("Could not create file")
	}

	err = fs.AddLatency(filename, latency)
	if err != nil {
		t.Error("Could not add latency to file operations")
	}

	fileLatency, err := fs.GetLatency(filename)
	if err != nil {
		t.Errorf("GetLatency returned error: %s", err)
	}
	if fileLatency != latency {
		t.Errorf("Returned file latency '%s' different than the one added '%s'", fileLatency, latency)
	}

	startTime := time.Now()
	fileInfo, err := fs.Stat(filename)
	duration := time.Since(startTime)
	if err != nil {
		t.Errorf("File read operation with latency returned error: %s", err)
	}
	if fileInfo.Name() != filename {
		t.Errorf("File stat didn't return the correct filename: %s", fileInfo.Name())
	}

	if duration < latency {
		t.Error("File read operation didn't took more than the latency duration")
	}

}
