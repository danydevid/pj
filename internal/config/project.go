package config

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"
)

// FileLock holds the active file handle for the lock file.
type FileLock struct {
	file *os.File
}

// AcquireLock attempts to acquire an exclusive lock on the file with a 5-second timeout.
func AcquireLock(lockFilePath string) (*FileLock, error) {
	file, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		for {
			err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
			if err == nil {
				done <- nil
				return
			}
			if ctx.Err() != nil {
				done <- ctx.Err()
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	select {
	case err := <-done:
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("could not acquire lock within 5 seconds: %w", err)
		}
	case <-ctx.Done():
		file.Close()
		return nil, fmt.Errorf("lock acquisition timed out (5s)")
	}

	return &FileLock{file: file}, nil
}

// Unlock releases the file lock and closes the handle.
func (fl *FileLock) Unlock() error {
	if fl.file == nil {
		return nil
	}
	_ = syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
	return fl.file.Close()
}
