package main

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/mattn/go-shellwords"
	"github.com/spf13/viper"
)

// Package contains basic information about a library.
type Package struct {
	Name string
}

func addPackage(options *CompilerOptions, name string) *Package {
	if ret := addPackageLocal(options, name); ret != nil {
		return ret
	}

	//TODO: Implement global packages

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if ret := addPackagePkgconfig(options, name); ret != nil {
			return ret
		}
	}

	//TODO: Implement vcpkg
	//if runtime.GOOS == "windows" {
	//}

	return nil
}

func addPackageLocal(options *CompilerOptions, name string) *Package {
	packageInfo := viper.GetStringMap("package." + name)
	if len(packageInfo) == 0 {
		return nil
	}

	return configurePackageFromConfig(options, packageInfo, name)
}

func addPackagePkgconfig(options *CompilerOptions, name string) *Package {
	// pkg-config must be installed for this to work
	_, err := exec.LookPath("pkg-config")
	if err != nil {
		return nil
	}

	cmdCflags := exec.Command("pkg-config", name, "--cflags")
	outputCflags, err := cmdCflags.CombinedOutput()
	if err != nil {
		return nil
	}

	cmdLibs := exec.Command("pkg-config", name, "--libs")
	outputLibs, err := cmdLibs.CombinedOutput()
	if err != nil {
		return nil
	}

	parseCflags, _ := shellwords.Parse(strings.Trim(string(outputCflags), "\r\n"))
	parseLibs, _ := shellwords.Parse(strings.Trim(string(outputLibs), "\r\n"))

	options.CompilerFlagsCXX = append(options.CompilerFlagsCXX, parseCflags...)
	options.LinkerFlags = append(options.LinkerFlags, parseLibs...)

	return &Package{
		Name: name,
	}
}

func configurePackageFromConfig(options *CompilerOptions, pkg map[string]interface{}, name string) *Package {
	maybeUnpack(&options.IncludeDirectories, pkg["includes"])
	maybeUnpack(&options.LinkDirectories, pkg["linkdirs"])
	maybeUnpack(&options.LinkLibraries, pkg["links"])
	maybeUnpack(&options.Defines, pkg["defines"])
	maybeUnpack(&options.CompilerFlagsCXX, pkg["cflags"])
	maybeUnpack(&options.LinkerFlags, pkg["lflags"])

	return &Package{
		Name: name,
	}
}

func maybeUnpack(dest *[]string, src interface{}) {
	if src == nil {
		return
	}

	for _, val := range src.([]interface{}) {
		(*dest) = append(*dest, val.(string))
	}
}
