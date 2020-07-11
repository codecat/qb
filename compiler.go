package main

// Compiler contains information about the compiler
type Compiler interface {
	Compile(path, objDir string) error
	Link(objDir, outPath string) error
}
