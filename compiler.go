package main

// LinkType specifies the output build type
type LinkType int

const (
	// LinkExe will create an executable application. Will add the ".exe" suffix on Windows.
	LinkExe LinkType = iota

	// LinkDll will create a dynamic library. Will add the ".dll" suffix on Windows, and the ".so" suffix on Linux.
	LinkDll

	// LinkLib will create a static library. Will add the ".lib" suffix on Windows, and the ".a" suffix on Linux.
	LinkLib
)

// Compiler contains information about the compiler
type Compiler interface {
	Compile(path, objDir string, options *CompilerOptions) error
	Link(objDir, outPath string, outType LinkType, options *CompilerOptions) (string, error)
}

// CompilerOptions contains options used for compiling and linking
type CompilerOptions struct {
	// Static sets whether to build a completely-static binary (eg. no dynamic link libraries are loaded from disk)
	Static bool
}
