package main

import (
	"runtime"

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

	if runtime.GOOS == "linux" {
		//TODO: Implement pkgconfig
	}

	if runtime.GOOS == "windows" {
		//TODO: Implement vcpkg
	}

	return nil
}

func addPackageLocal(options *CompilerOptions, name string) *Package {
	packageInfo := viper.GetStringMap("package." + name)
	if len(packageInfo) == 0 {
		return nil
	}

	return configurePackageFromConfig(options, packageInfo, name)
}

func configurePackageFromConfig(options *CompilerOptions, pkg map[string]interface{}, name string) *Package {
	maybeUnpack(&options.IncludeDirectories, pkg["includes"])
	maybeUnpack(&options.LinkDirectories, pkg["linkdirs"])
	maybeUnpack(&options.LinkLibraries, pkg["links"])
	maybeUnpack(&options.Defines, pkg["defines"])
	maybeUnpack(&options.CompilerFlags, pkg["cflags"])
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
