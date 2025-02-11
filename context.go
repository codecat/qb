package main

// Context contains all the build system states that have to be remembered.
type Context struct {
	// Name is the name of the project.
	Name string

	// Final binary type we want to link.
	Type LinkType

	// SourceFiles contains paths to all the .c and .cpp files that have to be compiled.
	SourceFiles []string

	// ObjectPath is the intermediate folder where object files should be stored.
	ObjectPath string

	// OutPath is the directory where the final binary is written to.
	OutPath string

	// Compiler is an abstract interface used for compiling and linking on multiple platforms.
	Compiler               Compiler
	CompilerErrors         int
	CompilerOptions        *CompilerOptions
	CompilerWorkerChannel  chan CompilerWorkerTask
	CompilerWorkerFinished chan int
}

// NewContext creates a new context with initial values.
func NewContext() (*Context, error) {
	compiler, err := getCompiler()
	if err != nil {
		return nil, err
	}

	return &Context{
		Compiler:        compiler,
		CompilerOptions: &CompilerOptions{},

		SourceFiles: make([]string, 0),
	}, nil
}
