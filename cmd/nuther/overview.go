package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"nuther/internal/config"
	"nuther/internal/smart"
	"nuther/internal/ui/styles"
	"nuther/internal/ui/views"
)

func runOverview(args []string) error {
	flags := flag.NewFlagSet("overview", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	jsonPath := flags.String("json", "-", "smartctl JSON file to render ('-' for stdin)")
	width := flags.Int("width", 120, "render width")
	height := flags.Int("height", 50, "render height")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 && *jsonPath == "-" {
		*jsonPath = flags.Arg(0)
	}

	var data []byte
	var err error
	if *jsonPath == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(*jsonPath)
	}
	if err != nil {
		return err
	}

	drive, err := smart.ParseSmartctlDriveJSON(data)
	if err != nil {
		return err
	}

	fmt.Print(views.RenderOverview(drive, 0, 0, *width, *height, styles.NewStyles(config.DefaultConfig())))
	fmt.Println()
	return nil
}
