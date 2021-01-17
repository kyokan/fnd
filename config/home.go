package config

import (
	"os"

	"github.com/pkg/errors"
)

func HomeDirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if !stat.IsDir() {
		return false, errors.New("home dir path exists, but is a file")
	}

	return true, nil
}

func EnsureHomeDir(path string) error {
	exists, err := HomeDirExists(path)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("home directory does not exist - try running fnd init")
	}
	return nil
}

func InitHomeDir(homePath string) error {
	err := os.MkdirAll(homePath, 0700)
	if err != nil {
		return err
	}

	identity := NewIdentity()
	if err := WriteIdentity(homePath, identity); err != nil {
		return err
	}
	if err := InitBlobsDir(homePath); err != nil {
		return err
	}
	if err := InitDBDir(homePath); err != nil {
		return err
	}
	return WriteDefaultConfigFile(homePath)
}
