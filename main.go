package main

import (
	"fmt"
	"io"
	"os"

	"github.com/MixTapeSoftware/inctui/internal/incusapi"
	"github.com/MixTapeSoftware/inctui/internal/ui"
	"github.com/charmbracelet/log"
)

func main() {
	closer, err := setupLogger()
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer closer.Close()

	server, err := incusapi.NewInstanceServer()
	if err != nil {
		fmt.Println("Error loading server", err)
	}
	if _, err := incusui.InstancesUI(server); err == nil {
		fmt.Println("error:", err)
	}
	os.Exit(1)
}

func setupLogger() (io.Closer, error) {
	if os.Getenv("DEBUG") == "" {
		log.SetOutput(io.Discard)
		return io.NopCloser(nil), nil
	}

	f, err := os.OpenFile("debug.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	log.SetOutput(f)
	log.SetLevel(log.DebugLevel)
	// adds the line number
	log.SetReportCaller(true)
	return f, nil
}
