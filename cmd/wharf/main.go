package main

import (
	"fmt"

	wharfcmd "github.com/iver-wharf/wharf-cmd"
)

func main() {
	version, err := wharfcmd.GetVersion()
	if err != nil {
		fmt.Println("Failed to load version:", err)
	}
	execute(version)
}
