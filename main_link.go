package main

import (
	"github.com/spf13/viper"
)

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
