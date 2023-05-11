//go:build noui

package ui

import (
	"errors"
	"io/fs"
)

var ErrNoFiles = errors.New("no embedded files")

func WalkAssets(walkFcn func(string, fs.DirEntry, error) error) error {
	return ErrNoFiles
}

func GetAssetFiles() []string {
	_ = WalkAssets(func(s string, de fs.DirEntry, e error) error {
		return nil
	})

	return nil
}

func GetAssets() fs.FS {
	return nil
}
