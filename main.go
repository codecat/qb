package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/codecat/go-libs/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type workerTask struct {
	path      string
	outputDir string
}

var compiler Compiler
var compilerErrors int

var workerChan chan workerTask
var workerFinished chan int

func compileWorker(num int) {
	for {
		task, ok := <-workerChan
		if !ok {
			break
		}

		fileForward := strings.Replace(task.path, "\\", "/", -1)
		log.Info("%s", fileForward)

		err := compiler.Compile(task.path, task.outputDir)
		if err != nil {
			log.Error("Failed to compile %s!", fileForward)
			log.Error("%s", err.Error())
			compilerErrors++
		}
	}

	workerFinished <- num
}

func main() {
	// Configure logging
	log.CurrentConfig.Category = false

	// Prepare possible command line flags
	pflag.String("name", "", "binary output name without the extension")
	pflag.String("type", "exe", "binary output type, either \"exe\", \"dll\", or \"lib\"")
	pflag.Parse()

	// Load a qb.toml file, if it exists
	viper.AddConfigPath(".")
	viper.SetConfigName("qb")
	viper.BindPFlags(pflag.CommandLine)
	err := viper.ReadInConfig()
	if err == nil {
		log.Info("Using build configuration file %s", filepath.Base(viper.ConfigFileUsed()))
	}

	// Find the compiler
	compiler, err = getCompiler()
	if err != nil {
		log.Fatal("Unable to get compiler: %s", err.Error())
		return
	}

	// Find all the source files to compile
	sourceFiles, err := getSourceFiles()
	if err != nil {
		log.Fatal("Unable to read directory: %s", err.Error())
		return
	}

	if len(sourceFiles) == 0 {
		log.Warn("No source files found!")
		return
	}

	// Make a temporary folder for .obj files
	pathTmp := filepath.Join(os.TempDir(), fmt.Sprintf("qb_%d", time.Now().Unix()))
	os.Mkdir(pathTmp, 0777)
	defer os.RemoveAll(pathTmp)

	// Prepare worker channel
	workerChan = make(chan workerTask)
	workerFinished = make(chan int)

	// Start compiler worker routines
	numWorkers := runtime.NumCPU()
	if len(sourceFiles) < numWorkers {
		numWorkers = len(sourceFiles)
	}
	for i := 0; i < numWorkers; i++ {
		go compileWorker(i)
	}

	// Begin compilation timer
	timeStart := time.Now()

	// Compile all the source files
	for _, file := range sourceFiles {
		dir := filepath.Dir(file)
		outputDir := filepath.Join(pathTmp, dir)

		err := os.MkdirAll(outputDir, 0777)
		if err != nil {
			log.Error("Unable to create output directory %s: %s", outputDir, err.Error())
			compilerErrors++
			continue
		}

		workerChan <- workerTask{
			path:      file,
			outputDir: outputDir,
		}
	}

	// Close the worker channel
	close(workerChan)

	// Wait for all workers to finish compiling
	for i := 0; i < numWorkers; i++ {
		<-workerFinished
	}

	// Measure the time that compilation took
	timeCompilation := time.Since(timeStart)

	// Stop if there were any compiler errors
	if compilerErrors > 0 {
		log.Fatal("😢 Compilation failed!")
		return
	}

	// Link
	linkType := LinkExe
	switch viper.GetString("type") {
	case "exe":
		linkType = LinkExe
	case "dll":
		linkType = LinkDll
	case "lib":
		linkType = LinkLib
	}

	name := viper.GetString("name")
	if name == "" {
		currentDir, _ := filepath.Abs(".")
		name = filepath.Base(currentDir)
	}

	// Begin link timer
	timeStart = time.Now()

	outPath, err := compiler.Link(pathTmp, name, linkType)
	if err != nil {
		log.Fatal("👎 Link failed: %s", err.Error())
		return
	}

	// Measure the time that linking took
	timeLinking := time.Since(timeStart)

	// Report succcess
	log.Info("💚 %s", outPath)
	log.Info("⏳ compile %v, link %v", timeCompilation, timeLinking)
}

/*
TODO:
- Make sure all builds are completely statically linked
- Keep a state of already compiled files so subsequent builds are faster
- A nice progress bar of compilation/link status
- Pkgconfig support
*/
