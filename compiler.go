package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codecat/go-libs/log"
	"github.com/spf13/viper"
)

// LinkType specifies the output build type.
type LinkType int

const (
	// LinkExe will create an executable application. Will add the ".exe" suffix on Windows.
	LinkExe LinkType = iota

	// LinkDll will create a dynamic library. Will add the ".dll" suffix on Windows, and the ".so" suffix on Linux.
	LinkDll

	// LinkLib will create a static library. Will add the ".lib" suffix on Windows, and the ".a" suffix on Linux.
	LinkLib
)

// Compiler contains information about the compiler.
type Compiler interface {
	Compile(path, objDir string, options *CompilerOptions) error
	Link(objDir, outPath string, outType LinkType, options *CompilerOptions) (string, error)
	Clean(name string)
}

// ExceptionType is the way that a compiler's runtime might handle exceptions.
type ExceptionType int

const (
	// ExceptionsStandard is the standard way of handling exceptions, and will perform stack unwinding.
	ExceptionsStandard ExceptionType = iota

	// ExceptionsAll is only supported on Windows, and will allow catching excptions such as access violations and integer divide by zero exceptions.
	ExceptionsAll

	// ExceptionsMinimal is only supported on Windows, and is similar to ExceptionsAll, except there is no stack unwinding.
	ExceptionsMinimal
)

// OptimizeType defines the compiler optimization type.
type OptimizeType int

const (
	// OptimizeDefault favors speed for release builds, and disables optimization for debug builds.
	OptimizeDefault OptimizeType = iota

	// OptimizeNone performs no optimization at all.
	OptimizeNone

	// OptimizeSize favors size over speed in optimization.
	OptimizeSize

	// OptimizeSpeed favors sped over size in optimization.
	OptimizeSpeed
)

// CompilerOptions contains options used for compiling and linking.
type CompilerOptions struct {
	// Static sets whether to build a completely-static binary (eg. no dynamic link libraries are loaded from disk).
	Static bool

	// Debug configurations will add debug symbols. This will create a pdb file on Windows, and embed debugging information on Linux.
	Debug bool

	// Verbose compiling means we'll print the actual compiler and linker commands being executed.
	Verbose bool

	// Strict sets whether to be more strict on warnings.
	Strict bool

	// Include paths and library links
	IncludeDirectories []string
	LinkDirectories    []string
	LinkLibraries      []string

	// Additional compiler defines
	Defines []string

	// Additional compiler and linker flags
	CompilerFlags []string
	LinkerFlags   []string

	// Specific options
	Exceptions   ExceptionType
	Optimization OptimizeType
}

// CompilerWorkerTask describes a task for the compiler worker
type CompilerWorkerTask struct {
	path      string
	outputDir string
}

func compileWorker(ctx *Context, num int) {
	for {
		// Get a task
		task, ok := <-ctx.CompilerWorkerChannel
		if !ok {
			break
		}

		// Log the file we're currently compiling
		fileForward := strings.Replace(task.path, "\\", "/", -1)
		log.Info("%s", fileForward)

		// Invoke the compiler
		err := ctx.Compiler.Compile(task.path, task.outputDir, ctx.CompilerOptions)
		if err != nil {
			log.Error("Failed to compile %s!\n%s", fileForward, err.Error())
			ctx.CompilerErrors++
		}
	}

	ctx.CompilerWorkerFinished <- num
}

func performCompilation(ctx *Context) {
	// Prepare worker channels
	ctx.CompilerWorkerChannel = make(chan CompilerWorkerTask)
	ctx.CompilerWorkerFinished = make(chan int)

	// Start compiler worker routines
	numWorkers := runtime.NumCPU()
	if len(ctx.SourceFiles) < numWorkers {
		numWorkers = len(ctx.SourceFiles)
	}
	for i := 0; i < numWorkers; i++ {
		go compileWorker(ctx, i)
	}

	// Compile all the source files
	for _, file := range ctx.SourceFiles {
		// The output dir will be a sub-folder in the object directory
		dir := filepath.Dir(file)
		outputDir := filepath.Join(ctx.ObjectPath, dir)

		err := os.MkdirAll(outputDir, 0777)
		if err != nil {
			log.Error("Unable to create output directory %s: %s", outputDir, err.Error())
			ctx.CompilerErrors++
			continue
		}

		// Send the task to an available worker
		ctx.CompilerWorkerChannel <- CompilerWorkerTask{
			path:      file,
			outputDir: outputDir,
		}
	}

	// Close the worker channel
	close(ctx.CompilerWorkerChannel)

	// Wait for all workers to finish compiling
	for i := 0; i < numWorkers; i++ {
		<-ctx.CompilerWorkerFinished
	}
}

func performLinking(ctx *Context) (string, error) {
	// Get the link type
	linkType := LinkExe
	switch viper.GetString("type") {
	case "exe":
		linkType = LinkExe
	case "dll":
		linkType = LinkDll
	case "lib":
		linkType = LinkLib
	}

	// Invoke the linker
	return ctx.Compiler.Link(ctx.ObjectPath, ctx.Name, linkType, ctx.CompilerOptions)
}
