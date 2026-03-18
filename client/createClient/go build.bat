@echo off
go build -ldflags "-s -w -H=windowsgui -extldflags '-static'" -o loader.exe
pause