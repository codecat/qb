// +build linux

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type linuxCompiler struct {
}

func (ci linuxCompiler) Compile(path, objDir string) error {
	filename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	args := make([]string, 0)
	args = append(args, "-c")
	args = append(args, "-o", filepath.Join(objDir, filename+".o"))
	args = append(args, path)

	cmd := exec.Command("gcc", args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.Trim(string(outputBytes), "\r\n")
		return errors.New(output)
	}
	return nil
}

func (ci linuxCompiler) Link(objDir, outPath string, outType LinkType) (string, error) {
	args := make([]string, 0)

	switch outType {
	case LinkExe:
	case LinkDll:
		outPath += ".so"
		args = append(args, "-shared")
	case LinkLib:
		outPath += ".a"
		args = append(args, "-static")
	}

	args = append(args, "-o", outPath)
	args = append(args, "-std=c++17")
	args = append(args, "-static-libgcc")
	args = append(args, "-static-libstdc++")
	//args = append(args, "-lm")

	filepath.Walk(objDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".o") {
			return nil
		}
		args = append(args, path)
		return nil
	})

	cmd := exec.Command("gcc", args...)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.Trim(string(outputBytes), "\r\n")
		return "", errors.New(output)
	}
	return outPath, nil
}
