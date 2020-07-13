// +build linux

package main

import (
	"errors"
	"os/exec"
)

func getCompiler() (Compiler, error) {
	//TODO: Check if we have gcc and ld installed
	//TODO: Add option for other flavors of gcc (eg. mingw)

	toolset := ""

	if _, err := exec.LookPath("clang"); err == nil {
		toolset = "clang"
	} else if _, err := exec.LookPath("gcc"); err == nil {
		toolset = "gcc"
	} else {
		return nil, errors.New("couldn't find clang or gcc in the PATH")
	}

	return linuxCompiler{
		toolset: toolset,
	}, nil
}
