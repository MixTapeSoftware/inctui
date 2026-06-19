package main

import (
	"fmt"
	"os"

	"github.com/MixTapeSoftware/inctui/internal/incusapi"
	"github.com/MixTapeSoftware/inctui/internal/ui"
)

func main() {
	fetcher, err := incusapi.NewInstanceFetcher()

	if err != nil {
		fmt.Println("Error loading fetcher", err)
	}
	if _, err := incusui.InstancesUI(fetcher); err == nil {
		fmt.Println("error:", err)
	}
	os.Exit(1)
}
