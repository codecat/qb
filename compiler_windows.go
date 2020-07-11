// +build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func (ci windowsCompiler) Compile(path, objDir string) error {
	// cl.exe args: https://docs.microsoft.com/en-us/cpp/build/reference/compiler-options-listed-by-category?view=vs-2019

	filename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	args := make([]string, 0)
	args = append(args, "/nologo")
	args = append(args, "/c")
	args = append(args, fmt.Sprintf("/Fo%s\\%s.obj", objDir, filename))
	args = append(args, path)

	cmd := exec.Command(ci.compiler(), args...)
	cmd.Env = append(os.Environ(),
		"INCLUDE="+strings.Join(ci.includeDirs(), ";"),
	)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.Trim(string(outputBytes), "\r\n")
		return errors.New(output)
	}
	return nil
}

func (ci windowsCompiler) Link(objDir, outPath string) error {
	// link.exe args: https://docs.microsoft.com/en-us/cpp/build/reference/linker-options?view=vs-2019

	args := make([]string, 0)
	args = append(args, "/nologo")
	//args = append(args, "/dll")
	args = append(args, "/machine:x64")
	args = append(args, fmt.Sprintf("/out:%s", outPath))

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

	cmd := exec.Command(ci.linker(), args...)
	cmd.Env = append(os.Environ(),
		"LIB="+strings.Join(ci.linkDirs(), ";"),
	)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.Trim(string(outputBytes), "\r\n")
		return errors.New(output)
	}
	return nil
}