package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codecat/go-libs/log"
)

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
