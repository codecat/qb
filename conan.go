package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
)

type Conanfile map[string][]string

func loadConanFile(path string) (Conanfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	regexHeader := regexp.MustCompile(`^\[(.*)\]$`)
	currentHeader := ""

	ret := make(Conanfile)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		res := regexHeader.FindStringSubmatch(line)
		if len(res) == 0 {
			if currentHeader == "" {
				return nil, fmt.Errorf("invalid conanfile on line: \"%s\"", line)
			}
			ret[currentHeader] = append(ret[currentHeader], strings.Trim(line, " \t"))
		} else {
			currentHeader = res[1]
		}
	}

	return ret, nil
}

func addConanPackages(ctx *Context, conan Conanfile) {
	//TODO: Implement all Conan features
	// conan["frameworkdirs"] // contains .framework files, only needed on MacOS
	// conan["frameworks"] // frameworks to link to
	// conan["bindirs"] // contains .dll files, only needed when linking with shared libraries

	isWindows := runtime.GOOS == "windows"

	// contains .h files
	for _, dir := range conan["includedirs"] {
		ctx.CompilerOptions.IncludeDirectories = append(ctx.CompilerOptions.IncludeDirectories, dir)
	}

	// contains .lib files
	for _, dir := range conan["libdirs"] {
		ctx.CompilerOptions.LinkDirectories = append(ctx.CompilerOptions.LinkDirectories, dir)
	}

	// libraries to link to
	for _, lib := range conan["libs"] {
		if isWindows && !strings.HasSuffix(lib, ".lib") {
			lib += ".lib"
		}
		ctx.CompilerOptions.LinkLibraries = append(ctx.CompilerOptions.LinkLibraries, lib)
	}

	// additional system libraries to link to
	for _, lib := range conan["system_libs"] {
		if isWindows && !strings.HasSuffix(lib, ".lib") {
			lib += ".lib"
		}
		ctx.CompilerOptions.LinkLibraries = append(ctx.CompilerOptions.LinkLibraries, lib)
	}

	// precompiler defines to add
	for _, define := range conan["defines"] {
		ctx.CompilerOptions.Defines = append(ctx.CompilerOptions.Defines, define)
	}

	// C++ compiler flags to add
	for _, flag := range conan["cppflags"] {
		ctx.CompilerOptions.CompilerFlagsCPP = append(ctx.CompilerOptions.CompilerFlagsCPP, flag)
	}

	// C/C++ compiler flags to add
	for _, flag := range conan["cxxflags"] {
		ctx.CompilerOptions.CompilerFlagsCXX = append(ctx.CompilerOptions.CompilerFlagsCXX, flag)
	}

	// C compiler flags to add
	for _, flag := range conan["cflags"] {
		ctx.CompilerOptions.CompilerFlagsC = append(ctx.CompilerOptions.CompilerFlagsC, flag)
	}

	if ctx.Type == LinkDll {
		// linker flags to add when building a shared library
		for _, flag := range conan["sharedlinkflags"] {
			ctx.CompilerOptions.LinkerFlags = append(ctx.CompilerOptions.LinkerFlags, flag)
		}

	} else if ctx.Type == LinkExe {
		// linker flags to add when building an executable
		for _, flag := range conan["exelinkflags"] {
			ctx.CompilerOptions.LinkerFlags = append(ctx.CompilerOptions.LinkerFlags, flag)
		}
	}
}
