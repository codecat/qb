// +build darwin

package main

func getCompiler() (Compiler, error) {
	return darwinCompiler{}, nil
}
