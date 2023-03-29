package ffuncs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

type writeCloseSyncer interface {
	io.ReadWriteCloser
	Sync() error
}

func CreateFile(args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.Errorf("create file: invalid count args ext: 1 got: %d", len(args))
	}

	filePath := args[0]
	wc, err := createFile(filePath)
	if err != nil {
		return "", err
	}

	return filePath, wc.Close()
}

func RemoveFile(args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.Errorf("remove file: invalid count args ext: 1 got: %d", len(args))
	}
	filePath := args[0]
	return "", os.Remove(filePath)
}

func CTimeFile(args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.Errorf("ctime file: invalid count args ext: 1 got: %d", len(args))
	}

	filePath := args[0]
	info, isExists := fileIsExists(filePath)
	if !isExists {
		return "", errors.Wrapf(os.ErrNotExist, "get ctime file %s error", filePath)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if ok {
		ctime := time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
		return ctime.Format(TimeFormat), nil
	}
	return "", errors.New("failed get ctime")
}

func RenameFile(args []string) (string, error) {
	if len(args) != 2 {
		return "", errors.Errorf("rename file: invalid count args ext: 2 got: %d", len(args))
	}
	oldPath, newPath := args[0], args[1]
	return newPath, os.Rename(oldPath, newPath)
}

func ValidateCondition(args []string) (string, error) {
	if len(args) != 3 {
		return "", errors.Errorf("validate condition: invalid count args ext: 3 got: %d", len(args))
	}
	t1, err := time.Parse(TimeFormat, args[0])
	if err != nil {
		return "", err
	}
	t2, err := time.Parse(TimeFormat, args[1])
	if err != nil {
		return "", err
	}
	filePath := args[2]
	if t1.After(t2) {
		return filePath, nil
	}
	return "", ErrFalseCondition
}

func WriteString(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.Errorf("write to file: args is empty")
	}
	// first args filename
	filePath := args[0]
	// strings to write file
	writeStrings := args[1:]
	if _, isExist := fileIsExists(filePath); !isExist {
		return "", errors.Wrapf(os.ErrNotExist, "write to file %s error", filePath)
	}

	wcSync, err := createFile(filePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	for _, str := range writeStrings {
		buf.WriteString(str)
		buf.WriteByte('\n')
	}

	_, err = io.Copy(wcSync, &buf)
	if err != nil {
		return "", errors.Wrapf(err, "copy to file %s error", filePath)
	}
	if err = wcSync.Sync(); err != nil {
		return "", errors.Wrapf(err, "sync file %s error", filePath)
	}

	return filePath, wcSync.Close()
}
func fileIsExists(path string) (os.FileInfo, bool) {
	var info, err = os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false
	}
	return info, !info.IsDir()
}

func createFile(path string) (writeCloseSyncer, error) {
	const (
		flag = os.O_WRONLY | os.O_CREATE
		perm = 0644
	)

	f, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("failed open file %s: %w", path, err)
	}
	return f, nil
}
