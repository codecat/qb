package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
	pflag.String("out", "", "the output directory")
	pflag.String("type", "exe", "binary output type, either \"exe\", \"dll\", or \"lib\"")
	pflag.Bool("static", false, "link statically to create a standalone binary")
	pflag.Bool("debug", false, "produce debug information")
	pflag.Bool("verbose", false, "print all compiler and linker commands being executed")
	pflag.Bool("strict", false, "be more strict in compiler warnings")
	pflag.String("exceptions", "std", "way to handle exceptions, either \"std\", \"all\", or \"min\"")
	pflag.String("optimize", "default", "enable optimizations, either \"defualt\", \"none\", \"size\", or \"speed\"")
	pflag.String("cppstd", "latest", "select the C++ standard to use, either \"latest\", \"20\", \"17\", or \"14\"")
	pflag.String("cstd", "latest", "select the C standard to use, either \"latest\", \"17\", or \"11\"")
	pflag.StringSlice("include", nil, "directories to add to the include path")
	pflag.StringSlice("define", nil, "adds a precompiler definition")
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
		os.Exit(1)
	}

	// Find the name of the project
	ctx.Name = viper.GetString("name")
	if ctx.Name == "" {
		// If there's no name set, use the name of the current directory
		currentDir, _ := filepath.Abs(".")
		ctx.Name = filepath.Base(currentDir)
	}

	// Get the output path
	ctx.OutPath = viper.GetString("out")

	// Get the link type
	switch viper.GetString("type") {
	case "exe":
		ctx.Type = LinkExe
	case "dll":
		ctx.Type = LinkDll
	case "lib":
		ctx.Type = LinkLib
	default:
		ctx.Type = LinkExe
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
	ctx.CompilerOptions.Strict = viper.GetBool("strict")

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

	// Load optimization options
	optimizeType := viper.GetString("optimize")
	switch optimizeType {
	case "", "default":
		ctx.CompilerOptions.Optimization = OptimizeDefault
	case "none":
		ctx.CompilerOptions.Optimization = OptimizeNone
	case "size":
		ctx.CompilerOptions.Optimization = OptimizeSize
	case "speed":
		ctx.CompilerOptions.Optimization = OptimizeSpeed
	default:
		log.Warn("Unrecognized optimization type %s", optimizeType)
	}

	// If we're on default optimization, optimize for speed if we're not a debug build
	if ctx.CompilerOptions.Optimization == OptimizeDefault && !ctx.CompilerOptions.Debug {
		ctx.CompilerOptions.Optimization = OptimizeSpeed
	}

	// Load C++ compiler standard
	cppStandard := viper.GetString("cppstd")
	switch cppStandard {
	case "", "latest":
		ctx.CompilerOptions.CPPStandard = CPPStandardLatest
	case "20":
		ctx.CompilerOptions.CPPStandard = CPPStandard20
	case "17":
		ctx.CompilerOptions.CPPStandard = CPPStandard17
	case "14":
		ctx.CompilerOptions.CPPStandard = CPPStandard14
	default:
		log.Warn("Unrecognized C++ compiler standard %s", cppStandard)
	}

	// Load C compiler standard
	cStandard := viper.GetString("cstd")
	switch cStandard {
	case "", "latest":
		ctx.CompilerOptions.CStandard = CStandardLatest
	case "17":
		ctx.CompilerOptions.CStandard = CStandard17
	case "11":
		ctx.CompilerOptions.CStandard = CStandard11
	default:
		log.Warn("Unrecognized C compiler standard %s", cStandard)
	}

	// Add custom include directories
	includes := viper.GetStringSlice("include")
	for _, include := range includes {
		fi, err := os.Stat(include)
		if err != nil {
			log.Warn("Unable to include directory %s: %s", include, err.Error())
			continue
		}

		if !fi.IsDir() {
			log.Warn("Include path is not a directory: %s", include)
			continue
		}

		ctx.CompilerOptions.IncludeDirectories = append(ctx.CompilerOptions.IncludeDirectories, include)
	}

	// Add preprocessor definitions
	defines := viper.GetStringSlice("define")
	ctx.CompilerOptions.Defines = append(ctx.CompilerOptions.Defines, defines...)

	// Find packages
	packages := viper.GetStringSlice("pkg")
	for _, pkg := range packages {
		pkgInfo := addPackage(ctx.CompilerOptions, pkg)
		if pkgInfo == nil {
			log.Warn("Unable to find package %s!", pkg)
			continue
		}
	}

	// To support Conan: run "conan install", if a conanfile exists, but conanbuildinfo.txt does not exist
	if fileExists("conanfile.txt") && !fileExists("conanbuildinfo.txt") {
		log.Info("Conanfile found: installing dependencies from Conan")
		err := exec.Command("conan", "install", ".").Run()
		if err != nil {
			log.Warn("Conan install failed: %s", err.Error())
		}
	}

	// To support Conan: use conanbuildinfo.txt, if it exists
	if fileExists("conanbuildinfo.txt") {
		conan, err := loadConanFile("conanbuildinfo.txt")
		if err != nil {
			log.Warn("Unable to load conanbuildinfo.txt: %s", err.Error())
		} else {
			addConanPackages(ctx, conan)
		}
	}

	// Find all the source files to compile
	ctx.SourceFiles, err = getSourceFiles()
	if err != nil {
		log.Fatal("Unable to read directory: %s", err.Error())
		os.Exit(1)
	}

	if len(ctx.SourceFiles) == 0 {
		log.Warn("No source files found!")
		os.Exit(1)
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
		os.Exit(1)
	}

	// Perform the linking
	timeStart = time.Now()
	outPath, err := performLinking(ctx)
	timeLinking := time.Since(timeStart)

	// Stop if linking failed
	if err != nil {
		log.Fatal("😢 Link failed!")
		log.Fatal("%s", err.Error())
		os.Exit(1)
	}

	// Report succcess
	log.Info("👏 %s", outPath)
	log.Info("⏳ compile %v, link %v", timeCompilation, timeLinking)

	// Run the binary if it's requested
	if hasCommand("run") && viper.GetString("type") == "exe" {
		// We have to use the absolute path here to make this work on Linux and MacOS
		absOutPath, _ := filepath.Abs(outPath)
		cmd := exec.Command(absOutPath)
		cmd.Args = slices.Concat([]string{absOutPath}, pflag.Args()[1:])
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}
