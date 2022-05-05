package badfs

import (
	"errors"
	"io/fs"
	"os"
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
	defer fs.DelWriteError(symlinkName)

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

func TestBadFsReadlinkIfPossible(t *testing.T) {
	const readErrDesc = "read file error"
	// MemMapFs does not support symlinks
	memFs := New(afero.NewMemMapFs())
	osFs := New(afero.NewOsFs())

	_, err := memFs.ReadlinkIfPossible("anyname")

	if err == nil {
		t.Error("Wrapped MemMapFs should return error on ReadlinkifPossible")
	}

	file, err := afero.TempFile(osFs, "", "afero")

	if err != nil {
		t.Errorf("Unable to create temp file: %s", err)
	}

	filename := file.Name()
	defer osFs.Remove(filename)

	// A crude and slow way to create a unique valid symlink name is to create a temp file and delete it
	// while keeping its name.
	symlink, err := afero.TempFile(osFs, "", "afero")

	if err != nil {
		t.Errorf("Unable to create symlink file: %s", err)
	}

	symlinkName := symlink.Name()
	err = osFs.Remove(symlinkName)
	if err != nil {
		t.Errorf("Unable to remove symlink temp file: %s", err)
	}

	// Create symlink
	err = osFs.SymlinkIfPossible(filename, symlinkName)

	if err != nil {
		t.Errorf("SymLinkIfPossible returned error: %s", err)
	}
	// Test clean up
	defer osFs.Remove(symlinkName)

	linkName, err := osFs.ReadlinkIfPossible(symlinkName)

	if err != nil {
		t.Errorf("ReadlinkIfPossible returned error: %s", err)
	}

	if linkName != filename {
		t.Error("resolved link file name is different than temp test file name")
	}

	osFs.AddReadError(symlinkName, errors.New(readErrDesc))

	_, err = osFs.ReadlinkIfPossible(symlinkName)

	if err == nil {
		t.Errorf("ReadlinkIfPossible should have returned")
	}

	if err.Error() != readErrDesc {
		t.Error("Read error text is different than the one configured for symlink")
	}

}

func TestBadRename(t *testing.T) {
	const filename1 = "myTestFile1"
	const filename2 = "myTestFile2"
	const writeErrDesc = "write file error"
	fs := New(afero.NewMemMapFs())

	_, err := fs.Create(filename1)

	if err != nil {
		t.Error("Could not create test file")
	}

	err = fs.Rename(filename1, filename2)

	if err != nil {
		t.Error("Could not rename test file")
	}

	fs.AddWriteError(filename2, errors.New(writeErrDesc))

	err = fs.Rename(filename2, filename1)

	if err == nil {
		t.Error("Rename should have returned a write error")
	}

	if err.Error() != writeErrDesc {
		t.Error("Returned write error text does not match the one configured")
	}

}

func TestBadRemoveAll(t *testing.T) {
	const filename = "./full/path/filename"
	const writeErrDesc = "write file error"
	const errorDir = "./full/path"
	const inexistentPath = "/inexistent/path"

	fs := New(afero.NewMemMapFs())

	_, err := fs.Create(filename)

	if err != nil {
		t.Error("Could not create test file")
	}

	err = fs.RemoveAll("./full")

	if err != nil {
		t.Errorf("RemoveAll returned error: %s", err)
	}

	//fs.Open(filename)

	// Create it again
	_, err = fs.Create(filename)

	if err != nil {
		t.Errorf("Could not create test file: %s", err)
	}

	fs.AddWriteError(errorDir, errors.New(writeErrDesc))

	// Try removing but now with full path to file

	err = fs.RemoveAll("./full")

	if err == nil {
		t.Error("RemoveAll should have returned a write error")
	}

	err = fs.RemoveAll(inexistentPath)
	// MemMapFs does not return error on inexistent directory paths nor should BadFs

	if err != nil {
		t.Error("RemoveAll returned error for inexistent path")
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

func TestBadOpenFile(t *testing.T) {
	const errorFreeFilename = "myTestFile"
	const writeErrorFilename = "myWETestFile"
	const readErrorFilename = "myRETestFile"
	const mode fs.FileMode = 0666
	const writeErrDesc = "write file error"
	const readErrDesc = "read file error"
	fs := New(afero.NewMemMapFs())

	file, err := fs.OpenFile(errorFreeFilename, os.O_CREATE, mode)
	if err != nil {
		t.Errorf("Error creating test file with no errors: %s", err)
	}
	if file.Name() != errorFreeFilename {
		t.Error("File with no error name does not match the provided test filename")
	}

	fs.AddWriteError(writeErrorFilename, errors.New(writeErrDesc))

	_, err = fs.OpenFile(writeErrorFilename, os.O_CREATE, mode)

	if err == nil {
		t.Error("OpenFile should have returned a write error")
	}

	if err.Error() != writeErrDesc {
		t.Error("Write error text does not match the configured test error")
	}

	fs.AddReadError(readErrorFilename, errors.New(readErrDesc))

	_, err = fs.OpenFile(readErrorFilename, os.O_RDONLY, mode)

	if err == nil {
		t.Error("OpenFile should have returned a read error")
	}

	if err.Error() != readErrDesc {
		t.Error("Read error text does not match the configured test error")
	}
}

func TestBadOpen(t *testing.T) {
	const filename = "myTestFile"
	const readErrorText = "read error text"
	fs := New(afero.NewMemMapFs())

	_, err := fs.Create(filename)

	if err != nil {
		t.Errorf("Error creating test file with no errors: %s", err)
	}

	testFile, err := fs.Open(filename)

	if err != nil {
		t.Errorf("Could not open test file: %s", err)
	}

	if testFile.Name() != filename {
		t.Error("Opened test file name does not match the one created")
	}

	fs.AddReadError(filename, errors.New(readErrorText))

	_, err = fs.Open(filename)

	if err == nil {
		t.Errorf("Open failed to return read error for test file")
	}
}

func TestBadMkdir(t *testing.T) {
	const dirname = "directory"
	const writeErrDirname = "WEDirectory"
	const writeErrorText = "write error text"
	const mode os.FileMode = 0755

	fs := New(afero.NewMemMapFs())

	if err := fs.Mkdir(dirname, mode); err != nil {
		t.Error("Could not create a new directory")
	}

	info, err := fs.Stat(dirname)

	if err != nil {
		t.Errorf("Stat returned error: %s", err)
	}

	if info.Name() != dirname {
		t.Error("Returned test directory name does not match the created test directory name")
	}

	if !info.IsDir() {
		t.Error("Returned FileInfo does not have the directory flag set")
	}

	dirMode := info.Mode()
	if dirMode.Perm() != mode.Perm() {
		t.Error("Returned directory permissions does not match the one created")
	}

	fs.AddWriteError(writeErrDirname, errors.New(writeErrorText))

	if err := fs.Mkdir(writeErrDirname, mode); err == nil {
		t.Error("Mkdir did not return a write error")
	}

}
