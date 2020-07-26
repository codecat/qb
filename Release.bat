@echo off

rmdir /s /q Release

set GOARCH=amd64

set GOOS=windows
go build -ldflags "-s" -o Release/Windows/qb.exe

set GOOS=linux
go build -ldflags "-s" -o Release/Linux/qb

set GOOS=darwin
go build -ldflags "-s" -o Release/MacOS/qb
