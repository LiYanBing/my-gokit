package gokit_tool

import (
	"fmt"
	"path/filepath"
)

const (
	configTemplate = `{
	"address":":9090"
}`
)

func GenConfig(projectPath, pkgName string) error {
	path := filepath.Join(projectPath, "conf")
	err := createFile(filepath.Join(path, fmt.Sprintf("%s-local.conf", pkgName)), configTemplate, 0666)
	if err != nil {
		return err
	}

	err = createFile(filepath.Join(path, fmt.Sprintf("%s-test.conf", pkgName)), configTemplate, 0666)
	if err != nil {
		return err
	}

	err = createFile(filepath.Join(path, fmt.Sprintf("%s-pro.conf", pkgName)), configTemplate, 0666)
	if err != nil {
		return err
	}

	return nil
}
