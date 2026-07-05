package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"nuther/internal/config"
	"nuther/internal/screenshot"
	"nuther/internal/smart"
	"nuther/internal/ui/styles"
	"nuther/internal/ui/views"
)

func runOverview(args []string) error {
	flags := flag.NewFlagSet("overview", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	jsonPath := flags.String("json", "-", "smartctl JSON file to render ('-' for stdin)")
	outputPath := flags.String("output", "", "write overview image to this .png/.jpg path")
	themeName := flags.String("theme", "default", "theme to use for rendering")
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

	cfg := config.DefaultConfig()
	cfg.Theme = *themeName
	cfg.Colors = config.GetTheme(*themeName).Colors
	overview := views.RenderOverview(drive, 0, 0, *width, *height, styles.NewStyles(cfg))
	if *outputPath != "" {
		if err := screenshot.RenderOverviewImage(overview, *outputPath, cfg); err != nil {
			return err
		}
		fmt.Println(*outputPath)
		return nil
	}
	fmt.Print(overview)
	fmt.Println()
	return nil
}
