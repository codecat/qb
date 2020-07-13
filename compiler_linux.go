// +build linux

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codecat/go-libs/log"
)

type linuxCompiler struct {
	toolset string
}

func (ci linuxCompiler) Compile(path, objDir string, options *CompilerOptions) error {
	filename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	args := make([]string, 0)
	args = append(args, "-c")
	args = append(args, "-o", filepath.Join(objDir, filename+".o"))

	if options.Debug {
		args = append(args, "-g")
	}

	// Add include directories
	for _, dir := range options.IncludeDirectories {
		args = append(args, "-I"+dir)
	}

	// Add precompiler definitions
	for _, define := range options.Defines {
		args = append(args, "-D"+define)
	}

	// Add additional compiler flags
	for _, flag := range options.CompilerFlags {
		args = append(args, flag)
	}

	args = append(args, path)

	cmd := exec.Command(ci.toolset, args...)

	if options.Verbose {
		log.Trace("%s", strings.Join(cmd.Args, " "))
	}

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.Trim(string(outputBytes), "\r\n")
		return errors.New(output)
	}
	return nil
}

func (ci linuxCompiler) Link(objDir, outPath string, outType LinkType, options *CompilerOptions) (string, error) {
	args := make([]string, 0)

	exeName := ci.toolset

	switch outType {
	case LinkExe:
	case LinkDll:
		outPath += ".so"
		args = append(args, "-shared")
	case LinkLib:
		exeName = "ar"
		outPath += ".a"
	}

	if outType == LinkLib {
		// r = insert with replacement
		// c = create new archie
		// s = write an index
		args = append(args, "rcs")
		args = append(args, outPath)

	} else {
		args = append(args, "-o", outPath)
		args = append(args, "-std=c++17")

		if ci.toolset == "gcc" {
			args = append(args, "-static-libgcc")
			args = append(args, "-static-libstdc++")
		}

		if options.Static {
			args = append(args, "-static")
		}

		// Add additional library paths
		for _, dir := range options.LinkDirectories {
			args = append(args, "-L"+dir)
		}
	}

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

	if outType != LinkLib {
		// Link to some common standard libraries
		args = append(args, "-lstdc++")

		// Add libraries to link
		for _, link := range options.LinkLibraries {
			args = append(args, "-l"+link)
		}

		// Add additional linker flags
		for _, flag := range options.LinkerFlags {
			args = append(args, flag)
		}
	}

	cmd := exec.Command(exeName, args...)

	if options.Verbose {
		log.Trace("%s", strings.Join(cmd.Args, " "))
	}

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.Trim(string(outputBytes), "\r\n")
		return "", errors.New(output)
	}
	return outPath, nil
}

func (ci linuxCompiler) Clean(name string) {
	os.Remove(name)
	os.Remove(name + ".so")
	os.Remove(name + ".a")
}
