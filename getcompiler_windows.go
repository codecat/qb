// +build windows

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type vswhereOutput struct {
	InstanceID  string `json:"instanceId"`
	InstallPath string `json:"installationPath"`
}

func getCompiler() (Compiler, error) {
	ret := windowsCompiler{}

	vswherePath := "C:\\Program Files (x86)\\Microsoft Visual Studio\\Installer\\vswhere.exe"
	if !fileExists(vswherePath) {
		return nil, errors.New("couldn't find vswhere in the default path")
	}

	cmd := exec.Command(vswherePath, "-latest", "-format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var outputs []vswhereOutput
	json.Unmarshal(output, &outputs)

	if len(outputs) == 0 {
		return nil, errors.New("vswhere didn't return any installations")
	}

	ret.installDir = outputs[0].InstallPath
	installVersionBytes, err := ioutil.ReadFile(filepath.Join(ret.installDir, "VC\\Auxiliary\\Build\\Microsoft.VCToolsVersion.default.txt"))
	if err != nil {
		return nil, errors.New("unable to open Microsoft.VCToolsVersion.default.txt")
	}

	ret.installVersion = strings.Trim(string(installVersionBytes), "\r\n")

	// Get Windows 10 SDK path
	// HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\Microsoft\Microsoft SDKs\Windows\v10.0
	keySDK, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\WOW6432Node\\Microsoft\\Microsoft SDKs\\Windows\\v10.0", registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}

	ret.sdkDir, _, _ = keySDK.GetStringValue("InstallationFolder")
	ret.sdkVersion, _, _ = keySDK.GetStringValue("ProductVersion")

	return ret, nil
}
