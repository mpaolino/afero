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

// Test afero Fs interface

func TestBadFsChtimes(t *testing.T) {
	const filename = "myTestFile"
	const writeErrDesc = "write file error"
	fs := New(afero.NewMemMapFs())

	_, err := fs.Create(filename)

	if err != nil {
		t.Error("Could not create test file")
	}
	now := time.Now()
	chtime_err := fs.Chtimes(filename, now, now)

	if chtime_err != nil {
		t.Error("Could not change access and modify times in file test file")
	}

	fs.AddWriteError(filename, errors.New(writeErrDesc))

	chtime_err = fs.Chtimes(filename, now, now)
	if chtime_err == nil {
		t.Error("Chtimes should have returned a write error")
	}

	if chtime_err.Error() != writeErrDesc {
		t.Error("Chtimes write error description returned does not match the one added for the test file")
	}

}

func TestBadFsChmod(t *testing.T) {
	const filename = "myTestFile"
	const writeErrDesc = "write file error"
	fs := New(afero.NewMemMapFs())

	_, err := fs.Create(filename)

	if err != nil {
		t.Error("Could not create test file")
	}
	now := time.Now()
	chmod_err := fs.Chmod(filename, 0644)

	if chmod_err != nil {
		t.Error("Could not change file mode for test file")
	}

	fs.AddWriteError(filename, errors.New(writeErrDesc))

	chmod_err = fs.Chtimes(filename, now, now)

	if chmod_err == nil {
		t.Error("Chmod should have returned a write error")
	}

	if chmod_err.Error() != writeErrDesc {
		t.Error("Chmod write error description returned does not match the one added for the test file")
	}

}

func TestBadFsChown(t *testing.T) {
	const filename = "myTestFile"
	const writeErrDesc = "write file error"
	const uid = 1010
	const gid = 1010
	fs := New(afero.NewMemMapFs())

	_, err := fs.Create(filename)

	if err != nil {
		t.Error("Could not create test file")
	}

	chown_err := fs.Chown(filename, uid, gid)

	if chown_err != nil {
		t.Error("Could not change file owner for test file")
	}

	fs.AddWriteError(filename, errors.New(writeErrDesc))

	chown_err = fs.Chown(filename, uid, gid)

	if chown_err == nil {
		t.Error("Chown should have returned a write error")
	}

	if chown_err.Error() != writeErrDesc {
		t.Error("Chown write error text returned does not match the one added for the test file")
	}

}

func TestBadFsStat(t *testing.T) {
	const filename = "myTestFile"
	const readErrDesc = "read file error"

	fs := New(afero.NewMemMapFs())

	_, err := fs.Create(filename)

	if err != nil {
		t.Error("Could not create test file")
	}

	_, statErr := fs.Stat(filename)

	if statErr != nil {
		t.Error("Could get FileInfo with Stat for test file")
	}

	fs.AddReadError(filename, errors.New(readErrDesc))

	_, statErr = fs.Stat(filename)

	if statErr == nil {
		t.Error("Stat should have returned a read error")
	}

	if statErr.Error() != readErrDesc {
		t.Error("Stat read error text returned does not match the one added for the test file")
	}

}

func TestBadFsLstatIfPossible(t *testing.T) {
	const filename = "myTestFile"
	const readErrDesc = "read file error"

	fs := New(afero.NewMemMapFs())

	_, err := fs.Create(filename)

	if err != nil {
		t.Error("Could not create test file")
	}

	_, lsfOk, lstatErr := fs.LstatIfPossible(filename)

	if lstatErr != nil {
		t.Error("LstatIfPossible returned error for test file")
	}

	if lsfOk != false {
		t.Error("LstatIfPossible for memfs should have returned false")
	}

	fs.AddReadError(filename, errors.New(readErrDesc))

	_, _, lstatErr = fs.LstatIfPossible(filename)

	if lstatErr == nil {
		t.Error("LStatIfPossible should have returned a read error")
	}

	if lstatErr.Error() != readErrDesc {
		t.Error("LStatIfPossible read error text returned does not match the one added for the test file")
	}

}

func TestBadFsSymlinkIfPossible(t *testing.T) {
	const writeErrDesc = "write file error"
	const readErrDesc = "read file error"

	fs := New(afero.NewOsFs())

	file, err := afero.TempFile(fs, "", "afero")

	if err != nil {
		t.Errorf("Unable to create temp file: %s", err)
	}
	filename := file.Name()
	defer fs.Remove(filename)

	// A crude and slow way to create a unique valid symlink name is to create a temp file and delete it
	// while keeping its name.
	symlink, err := afero.TempFile(fs, "", "afero")
	if err != nil {
		t.Errorf("Unable to create symlink file: %s", err)
	}
	symlinkName := symlink.Name()
	err = fs.Remove(symlinkName)
	if err != nil {
		t.Errorf("Unable to remove symlink temp file: %s", err)
	}

	// Test clean up
	defer fs.Remove(symlinkName)
	defer fs.DelWriteError(filename)
	defer fs.DelReadError(filename)
	defer fs.DelWriteError(symlinkName)
	defer fs.DelReadError(symlinkName)

	if err != nil {
		t.Error("Could not create test file")
	}

	fs.AddReadError(filename, errors.New(readErrDesc))

	symLinkErr := fs.SymlinkIfPossible(filename, symlinkName)

	if symLinkErr != nil {
		t.Errorf("SymLinkIfPossible returned error: %s", symLinkErr)
	}

	symLinkErr = fs.checkReadError(symlinkName)

	if symLinkErr.Error() != readErrDesc {
		t.Error("SymLinkIfPossible did not copy the read error to the created symlink")
	}

	fs.AddWriteError(filename, errors.New(writeErrDesc))

	symLinkErr = fs.SymlinkIfPossible(filename, "newTestLink")

	if symLinkErr == nil {
		t.Error("SymlinkIfPossible should have returned a write error")
	}

	if symLinkErr.Error() != writeErrDesc {
		t.Error("SymlinkIfPossible write error text returned does not match the one added for the test file")
	}

}

func TestBadRemove(t *testing.T) {
	const filename = "myTestFile"
	const writeErrDesc = "write file error"
	fs := New(afero.NewMemMapFs())

	_, err := fs.Create(filename)

	if err != nil {
		t.Error("Could not create test file")
	}

	fs.AddWriteError(filename, errors.New(writeErrDesc))

	removeErr := fs.Remove(filename)

	if removeErr == nil {
		t.Error("Remove should have returned a write error")
	}

	if removeErr.Error() != writeErrDesc {
		t.Error("Remove write error text returned does not match the one added for the test file")
	}

	fs.DelWriteError(filename)

	removeErr = fs.Remove(filename)

	if removeErr != nil {
		t.Error("Remove returned a write error")
	}

	_, err = fs.GetWriteError(filename)
	if err == nil {
		t.Errorf("Fs should have returned and error getting the deleted write error for temp file, %s", err)
	}

}
