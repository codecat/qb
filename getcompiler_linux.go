// +build linux

package main

func getCompiler() (Compiler, error) {
	//TODO: Check if we have gcc and ld installed
	//TODO: Add option for other flavors of gcc (eg. mingw)

	return linuxCompiler{}, nil
}
