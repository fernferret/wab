//go:build !noui
// +build !noui

package ui

import (
	"embed"
	"io/fs"
)

// This module only serves to hod the embedded ui
// This is because go's embed refuses (for security reasons) to embed
// relative paths (which is fine). So I moved the ui.go over here!

//go:embed dist
var assets embed.FS

// WalkAssets is a shortcut function to walk all files in the assets embedded
// Filesystem
func WalkAssets(walkFcn func(string, fs.DirEntry, error) error) error {
	fsys, err := fs.Sub(assets, "dist")
	if err != nil {
		return err
	}

	return fs.WalkDir(fsys, ".", walkFcn)
}

// DumpAssetFiles is a helpful diagnostic function that prints all files in the
// embedded filesystem. Its used for tuning filepaths in the javascript
// generation code
func GetAssetFiles() []string {
	var files []string
	_ = WalkAssets(func(s string, de fs.DirEntry, e error) error {
		// Ignore the default "." dir. This is simply cosmetic as I don't want
		// to see the "."
		if s != "." {
			files = append(files, s)
		}

		return nil
	})

	return files
}

func GetAssets() fs.FS {
	fsys, err := fs.Sub(assets, "dist")
	if err != nil {
		panic(err)
	}

	return fsys
}
