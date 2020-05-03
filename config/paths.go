package config

import (
	"github.com/mitchellh/go-homedir"
	"os"
	"path"
)

const (
	BlobsPath = "blobs"
	DBPath    = "db"
)

func ExpandHomePath(path string) string {
	res, err := homedir.Expand(path)
	if err != nil {
		panic(err)
	}
	return res
}

func ExpandBlobsPath(homePath string) string {
	return path.Join(homePath, BlobsPath)
}

func InitBlobsDir(homePath string) error {
	p := ExpandBlobsPath(homePath)
	return os.MkdirAll(p, 0700)
}

func ExpandDBPath(homePath string) string {
	return path.Join(homePath, DBPath)
}

func InitDBDir(homePath string) error {
	p := ExpandDBPath(homePath)
	return os.MkdirAll(p, 0700)
}
