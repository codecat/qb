package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/codecat/go-libs/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func hasCommand(cmd string) bool {
	for _, arg := range pflag.Args() {
		if arg == cmd {
			return true
		}
	}
	return false
}

func main() {
	// Configure logging
	log.CurrentConfig.Category = false

	// Prepare possible command line flags
	pflag.String("name", "", "binary output name without the extension")
	pflag.String("type", "exe", "binary output type, either \"exe\", \"dll\", or \"lib\"")
	pflag.Bool("static", false, "link statically to create a standalone binary")
	pflag.Bool("debug", false, "produce debug information")
	pflag.Bool("verbose", false, "print all compiler and linker commands being executed")
	pflag.String("exceptions", "std", "way to handle exceptions, either \"std\", \"all\", or \"min\"")
	pflag.StringSlice("pkg", nil, "packages to link for compilation")
	pflag.Parse()

	// Load a qb.toml file, if it exists
	viper.AddConfigPath(".")
	viper.SetConfigName("qb")
	viper.BindPFlags(pflag.CommandLine)
	err := viper.ReadInConfig()
	if err == nil {
		log.Info("Using build configuration file %s", filepath.Base(viper.ConfigFileUsed()))
	}

	// Prepare qb's internal context
	ctx, err := NewContext()
	if err != nil {
		log.Fatal("Unable to initialize context: %s", err.Error())
		return
	}

	// Find the name of the project
	ctx.Name = viper.GetString("name")
	if ctx.Name == "" {
		// If there's no name set, use the name of the current directory
		currentDir, _ := filepath.Abs(".")
		ctx.Name = filepath.Base(currentDir)
	}

	// If we only have to clean, do that and exit
	if hasCommand("clean") {
		ctx.Compiler.Clean(ctx.Name)
		return
	}

	// Load any compiler options
	ctx.CompilerOptions.Static = viper.GetBool("static")
	ctx.CompilerOptions.Debug = viper.GetBool("debug")
	ctx.CompilerOptions.Verbose = viper.GetBool("verbose")

	// Load the exceptions method
	exceptionsType := viper.GetString("exceptions")
	switch exceptionsType {
	case "", "std", "standard":
		ctx.CompilerOptions.Exceptions = ExceptionsStandard
	case "all":
		ctx.CompilerOptions.Exceptions = ExceptionsAll
	case "min", "minimal":
		ctx.CompilerOptions.Exceptions = ExceptionsMinimal
	default:
		log.Warn("Unrecognized exceptions type %s", exceptionsType)
	}

	// Find packages
	packages := viper.GetStringSlice("pkg")
	for _, pkg := range packages {
		pkgInfo := addPackage(ctx.CompilerOptions, pkg)
		if pkgInfo == nil {
			log.Warn("Unable to find package %s!", pkg)
			continue
		}
	}

	// Find all the source files to compile
	ctx.SourceFiles, err = getSourceFiles()
	if err != nil {
		log.Fatal("Unable to read directory: %s", err.Error())
		return
	}

	if len(ctx.SourceFiles) == 0 {
		log.Warn("No source files found!")
		return
	}

	// Make a temporary folder for .obj files
	ctx.ObjectPath = filepath.Join(os.TempDir(), fmt.Sprintf("qb_%d", time.Now().Unix()))
	os.Mkdir(ctx.ObjectPath, 0777)
	defer os.RemoveAll(ctx.ObjectPath)

	// Perform the compilation
	timeStart := time.Now()
	performCompilation(ctx)
	timeCompilation := time.Since(timeStart)

	// Stop if there were any compiler errors
	if ctx.CompilerErrors > 0 {
		log.Fatal("😢 Compilation failed!")
		return
	}

	// Perform the linking
	timeStart = time.Now()
	outPath, err := performLinking(ctx)
	timeLinking := time.Since(timeStart)

	// Stop if linking failed
	if err != nil {
		log.Fatal("😢 Link failed!")
		log.Fatal("%s", err.Error())
		return
	}

	// Report succcess
	log.Info("👏 %s", outPath)
	log.Info("⏳ compile %v, link %v", timeCompilation, timeLinking)

	// Run the binary if it's requested
	if hasCommand("run") && viper.GetString("type") == "exe" {
		// We have to use the absolute path here to make this work on Linux and MacOS
		absOutPath, _ := filepath.Abs(outPath)
		cmd := exec.Command(absOutPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}
