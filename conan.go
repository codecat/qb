package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/codecat/go-libs/log"
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

	log.Info("Adding Conan packages")

	isWindows := runtime.GOOS == "windows"

	// contains .h files
	ctx.CompilerOptions.IncludeDirectories = append(ctx.CompilerOptions.IncludeDirectories, conan["includedirs"]...)

	// contains .lib files
	ctx.CompilerOptions.LinkDirectories = append(ctx.CompilerOptions.LinkDirectories, conan["libdirs"]...)

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
	ctx.CompilerOptions.Defines = append(ctx.CompilerOptions.Defines, conan["defines"]...)

	// C++ compiler flags to add
	ctx.CompilerOptions.CompilerFlagsCPP = append(ctx.CompilerOptions.CompilerFlagsCPP, conan["cppflags"]...)

	// C/C++ compiler flags to add
	ctx.CompilerOptions.CompilerFlagsCXX = append(ctx.CompilerOptions.CompilerFlagsCXX, conan["cxxflags"]...)

	// C compiler flags to add
	ctx.CompilerOptions.CompilerFlagsC = append(ctx.CompilerOptions.CompilerFlagsC, conan["cflags"]...)

	if ctx.Type == LinkDll {
		// linker flags to add when building a shared library
		ctx.CompilerOptions.LinkerFlags = append(ctx.CompilerOptions.LinkerFlags, conan["sharedlinkflags"]...)

	} else if ctx.Type == LinkExe {
		// linker flags to add when building an executable
		ctx.CompilerOptions.LinkerFlags = append(ctx.CompilerOptions.LinkerFlags, conan["exelinkflags"]...)
	}
}
