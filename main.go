package main

import (
	"fmt"
	"os"

	"github.com/MixTapeSoftware/inctui/internal/incusapi"
	"github.com/MixTapeSoftware/inctui/internal/ui"
)

func main() {
	server, err := incusapi.NewInstanceServer()

	if err != nil {
		fmt.Println("Error loading server", err)
	}
	if _, err := incusui.InstancesUI(server); err == nil {
		fmt.Println("error:", err)
	}
	os.Exit(1)
}
