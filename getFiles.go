package goaddtestcoverage

import (
	"os"
	"path/filepath"
)

// GetFiles takes a dir and returns two file lists. The first list is all the
// .js files in that dir, with init.js first in the list if it exists. The
// second is all the .js in the tests dir if it exists. Tests will also begin
// with init.js if it exists.
func GetFiles(dir string) ([]string, []string, error) {
	jsFiles, err := GetJSFiles(dir)
	if err != nil {
		return nil, nil, err
	}

	var testFiles []string
	testDir := filepath.Join(dir, "tests")
	ok, err := exists(testDir)
	if ok {
		testFiles, err = GetJSFiles(testDir)
	}
	if err != nil {
		return nil, nil, err
	}

	return jsFiles, testFiles, nil
}

// GetJSFiles returns all the .js files in a directory and will place init.js
// first in the list if it exists.
func GetJSFiles(dir string) ([]string, error) {
	jsFiles, err := filepath.Glob(filepath.Join(dir, "*.js"))
	if err != nil {
		return nil, err
	}

	for i, path := range jsFiles {
		if i > 0 {
			_, file := filepath.Split(path)
			if file == "init.js" {
				jsFiles[0], jsFiles[i] = jsFiles[i], jsFiles[0]
				break
			}
		}
	}

	return jsFiles, nil
}

func exists(dir string) (bool, error) {
	_, err := os.Stat(dir)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
