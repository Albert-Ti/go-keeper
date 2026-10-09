package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func scanDir(dir string) ([]localUploadedData, error) {
	var list []localUploadedData

	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	return list, filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}
		list = append(list, localUploadedData{
			filename: path,
			size:     fi.Size(),
		})
		return nil
	})
}
