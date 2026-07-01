package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	"nuther/internal/updater"
)

func runUpdate(args []string) error {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	flags.SetOutput(flag.CommandLine.Output())
	checkOnly := flags.Bool("check", false, "check for updates without installing")
	yes := flags.Bool("yes", false, "apply the update without an interactive confirmation prompt")
	repo := flags.String("repo", updater.DefaultRepo, "GitHub repository to update from")
	if err := flags.Parse(args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := updater.Client{Repo: *repo}
	info, err := client.Check(ctx, version)
	if errors.Is(err, updater.ErrNoUpdate) {
		fmt.Printf("nuther is up to date (%s)\n", version)
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Printf("Update available: %s -> %s\n", version, info.LatestVersion)
	fmt.Printf("Asset: %s\n", info.AssetName)
	fmt.Printf("Release: %s\n", info.ReleaseURL)
	if *checkOnly {
		return nil
	}

	if !*yes {
		fmt.Print("Apply update now? [y/N]: ")
		var answer string
		_, _ = fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" && answer != "yes" && answer != "YES" {
			fmt.Println("Update cancelled.")
			return nil
		}
	}

	if err := client.Apply(ctx, info); err != nil {
		return err
	}
	fmt.Printf("Updated to %s. Restart nuther to use the new version.\n", info.LatestVersion)
	return nil
}
