package main

import (
	"fmt"
	"runtime"
)

func GetArch(arch string) string {
	switch arch {
	case "amd64":
		{
			return "x64"
		}
	case "amd":
		{
			return "x86"
		}
	}

	return arch
}

func GetNodeFileVersion(node_version string) string {
	// get os info
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch os {
	case "windows":
		{
			// get the arch
			return fmt.Sprintf("node-v%s-win-%s.zip", node_version, GetArch(arch))
		}
	case "linux":
		{
			return ""
		}
	}

	panic("OS not supported currently")
}
