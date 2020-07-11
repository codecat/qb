package main

import (
	"os"
	"path/filepath"
	"regexp"
)

func getSourceFiles() ([]string, error) {
	ret := make([]string, 0)
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if ok, _ := regexp.Match("\\.(cpp|c)$", []byte(path)); !ok {
			return nil
		}

		ret = append(ret, path)
		return nil
	})
	return ret, err
}
