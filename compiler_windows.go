// +build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codecat/go-libs/log"
)

type windowsCompiler struct {
	installDir     string
	installVersion string

	sdkDir     string
	sdkVersion string
}

func (ci windowsCompiler) toolsDir() string {
	return filepath.Join(ci.installDir, "VC\\Tools\\MSVC", ci.installVersion)
}

func (ci windowsCompiler) binDir() string {
	return filepath.Join(ci.toolsDir(), "bin\\Hostx64\\x64")
}

func (ci windowsCompiler) sdkIncludeDir() string {
	return filepath.Join(ci.sdkDir, "include", ci.sdkVersion+".0")
}

func (ci windowsCompiler) sdkLibDir() string {
	return filepath.Join(ci.sdkDir, "lib", ci.sdkVersion+".0")
}

func (ci windowsCompiler) compiler() string {
	return filepath.Join(ci.binDir(), "cl.exe")
}

func (ci windowsCompiler) linker() string {
	return filepath.Join(ci.binDir(), "link.exe")
}

func (ci windowsCompiler) libber() string {
	return filepath.Join(ci.binDir(), "lib.exe")
}

func (ci windowsCompiler) includeDirs() []string {
	ret := make([]string, 0)

	// MSVC includes
	ret = append(ret, filepath.Join(ci.toolsDir(), "ATLMFC\\include"))
	ret = append(ret, filepath.Join(ci.toolsDir(), "include"))

	// Windows Kit includes
	ret = append(ret, filepath.Join(ci.sdkIncludeDir(), "ucrt"))
	ret = append(ret, filepath.Join(ci.sdkIncludeDir(), "shared"))
	ret = append(ret, filepath.Join(ci.sdkIncludeDir(), "um"))
	ret = append(ret, filepath.Join(ci.sdkIncludeDir(), "winrt"))
	ret = append(ret, filepath.Join(ci.sdkIncludeDir(), "cppwinrt"))

	return ret
}

func (ci windowsCompiler) linkDirs() []string {
	ret := make([]string, 0)

	// MSVC libraries
	ret = append(ret, filepath.Join(ci.toolsDir(), "ATLMFC\\lib\\x64"))
	ret = append(ret, filepath.Join(ci.toolsDir(), "lib\\x64"))

	// Windows Kit libraries
	ret = append(ret, filepath.Join(ci.sdkLibDir(), "ucrt\\x64"))
	ret = append(ret, filepath.Join(ci.sdkLibDir(), "um\\x64"))

	return ret
}

func (ci windowsCompiler) Compile(path, objDir string, options *CompilerOptions) error {
	// cl.exe args: https://docs.microsoft.com/en-us/cpp/build/reference/compiler-options-listed-by-category?view=vs-2019

	fileext := filepath.Ext(path)
	filename := strings.TrimSuffix(filepath.Base(path), fileext)

	args := make([]string, 0)
	args = append(args, "/nologo")
	args = append(args, "/c")

	// Set object output path
	args = append(args, fmt.Sprintf("/Fo%s\\%s.obj", objDir, filename))

	// Define the runtime flag
	runtimeFlag := "/M"
	if options.Static {
		runtimeFlag += "T"
	} else {
		runtimeFlag += "D"
	}
	if options.Debug {
		runtimeFlag += "d"
	}
	args = append(args, runtimeFlag)

	// Add exception handling flags
	if options.Exceptions == ExceptionsStandard {
		args = append(args, "/EHsc")
	} else if options.Exceptions == ExceptionsAll {
		args = append(args, "/EHa")
	} else if options.Exceptions == ExceptionsMinimal {
		args = append(args, "/EH")
	}

	// Add optimization flags
	if options.Optimization == OptimizeSize {
		args = append(args, "/O1")
	} else if options.Optimization == OptimizeSpeed {
		args = append(args, "/O2")
	}

	// Add include directories
	for _, dir := range options.IncludeDirectories {
		args = append(args, "/I"+dir)
	}

	// Add precompiler definitions
	for _, define := range options.Defines {
		args = append(args, "/D"+define)
	}

	// Add additional compiler flags
	for _, flag := range options.CompilerFlags {
		args = append(args, flag)
	}

	args = append(args, path)

	cmd := exec.Command(ci.compiler(), args...)
	cmd.Env = append(os.Environ(),
		"INCLUDE="+strings.Join(ci.includeDirs(), ";"),
	)

	if options.Verbose {
		log.Trace("%s", strings.Join(cmd.Args, " "))
	}

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.Trim(string(outputBytes), "\r\n")
		// Skip the first line from the output as it will always be the filename
		lines := strings.SplitN(output, "\n", 2)
		return errors.New(lines[1])
	}
	return nil
}

func (ci windowsCompiler) Link(objDir, outPath string, outType LinkType, options *CompilerOptions) (string, error) {
	// link.exe args: https://docs.microsoft.com/en-us/cpp/build/reference/linker-options?view=vs-2019

	exeName := ci.linker()

	args := make([]string, 0)
	args = append(args, "/nologo")
	args = append(args, "/machine:x64")
	args = append(args, "/incremental:no")

	if options.Debug {
		args = append(args, "/debug")
	}

	switch outType {
	case LinkExe:
		outPath += ".exe"

	case LinkDll:
		outPath += ".dll"
		args = append(args, "/dll")

	case LinkLib:
		exeName = ci.libber()
		outPath += ".lib"
		args = append(args, "/lib")
	}

	args = append(args, "/out:"+outPath)

	// Add additional library paths
	for _, dir := range options.LinkDirectories {
		args = append(args, "/libpath:"+dir)
	}

	// Link to some common standard libraries
	args = append(args, "kernel32.lib")
	args = append(args, "user32.lib")
	args = append(args, "advapi32.lib")

	// Add libraries to link
	for _, link := range options.LinkLibraries {
		args = append(args, link)
	}

	// Add additional linker flags
	for _, flag := range options.LinkerFlags {
		args = append(args, flag)
	}

	filepath.Walk(objDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".obj") {
			return nil
		}
		args = append(args, path)
		return nil
	})

	cmd := exec.Command(exeName, args...)
	cmd.Env = append(os.Environ(),
		"LIB="+strings.Join(ci.linkDirs(), ";"),
	)

	if options.Verbose {
		log.Trace("%s", strings.Join(cmd.Args, " "))
	}

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.Trim(string(outputBytes), "\r\n")
		return "", errors.New(output)
	}
	return outPath, nil
}

func (ci windowsCompiler) Clean(name string) {
	os.Remove(name + ".exe")
	os.Remove(name + ".dll")
	os.Remove(name + ".lib")
	os.Remove(name + ".pdb")
}
