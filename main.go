package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codecat/go-libs/log"
)

func main() {
	log.Open(log.CatTrace, log.CatFatal)

	// Find the compiler
	compiler, err := getCompiler()
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

	// Make a temporary folder for .obj files
	pathTmp := filepath.Join(os.TempDir(), fmt.Sprintf("qb_%d", time.Now().Unix()))
	os.Mkdir(pathTmp, 0777)
	defer os.RemoveAll(pathTmp)

	// Compile all the source files
	compileErrors := 0

	for _, file := range sourceFiles {
		fileForward := strings.Replace(file, "\\", "/", -1)
		log.Trace("%s", fileForward)

		dir := filepath.Dir(file)
		outputDir := filepath.Join(pathTmp, dir)

		err := os.MkdirAll(outputDir, 0777)
		if err != nil {
			log.Error("Unable to create output directory: %s", err.Error())
			compileErrors++
			continue
		}

		err = compiler.Compile(file, outputDir)
		if err != nil {
			log.Error("%s: %s", fileForward, err.Error())
			compileErrors++
			continue
		}
	}

	// Stop if there were any compiler errors
	if compileErrors > 0 {
		log.Fatal("Compiled failed: %d errors", compileErrors)
		return
	}

	// Link
	err = compiler.Link(pathTmp, "./Build.exe")
	if err != nil {
		log.Fatal("Link failed: %s", err.Error())
		return
	}

	// Report succcess
	log.Info("Build success!")
}

/*
TODO:
- Find output filename (.exe or .dll) from parent directory name
- Make sure all builds are completely statically linked
- Keep a state of already compiled files so we don't have to keep track of them
- Multiple tasks so we can compile multiple files at once
- A nice progress bar of compilation/link status
*/
