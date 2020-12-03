package gokit_tool

import (
	"fmt"
	"path/filepath"
)

const (
	configTemplate = `{
	"address":":4096"
}`
)

func GenConfig(projectPath, pkgName string) error {
	path := filepath.Join(projectPath, "conf")
	err := createFile(filepath.Join(path, fmt.Sprintf("%s-local.conf", pkgName)), configTemplate, false, 0666)
	if err != nil {
		return err
	}

	err = createFile(filepath.Join(path, fmt.Sprintf("%s-test.conf", pkgName)), configTemplate, false, 0666)
	if err != nil {
		return err
	}

	err = createFile(filepath.Join(path, fmt.Sprintf("%s-pro.conf", pkgName)), configTemplate, false, 0666)
	if err != nil {
		return err
	}

	return nil
}
