package main

import (
	"fmt"
	"os"

	incusui "github.com/MixTapeSoftware/inctui/internal/ui"
)

func main() {
	if _, err := incusui.InstancesUI(); err == nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
