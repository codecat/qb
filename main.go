package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/codecat/go-libs/log"
	"github.com/spf13/viper"
)

func main() {
	// Open logging
	log.Open(log.CatTrace, log.CatFatal)

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

	if len(sourceFiles) == 0 {
		log.Warn("No source files found!")
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

	outPath, err := compiler.Link(pathTmp, name, linkType)
	if err != nil {
		log.Fatal("Link failed: %s", err.Error())
		return
	}

	// Report succcess
	log.Info("Build success: %s", outPath)
}

/*
TODO:
- Make sure all builds are completely statically linked
- Keep a state of already compiled files so subsequent builds are faster
- Multiple tasks so we can compile multiple files at once
- A nice progress bar of compilation/link status
*/
