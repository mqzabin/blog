package extensions

import (
	"errors"
	"fmt"
	"os"
)

func EnsureSymlink(symlinkPath, relativeOriginalDirectory string) error {
	_, err := os.Stat(symlinkPath)
	switch {
	case err != nil && !errors.Is(err, os.ErrNotExist):
		return fmt.Errorf("reading symlink path file info: %w", err)
	case err == nil:
		return os.ErrExist
	}

	if err := os.Symlink(relativeOriginalDirectory, symlinkPath); err != nil {
		return err
	}

	return nil
}
