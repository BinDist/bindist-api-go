package main

import (
	"context"
	"fmt"
	"os"

	bindist "github.com/BinDist/bindist-api-go"
)

func main() {
	apiKey := os.Getenv("BINDIST_API_KEY")
	baseURL := os.Getenv("BINDIST_BASE_URL")

	if apiKey == "" {
		fmt.Println("BINDIST_API_KEY environment variable is required")
		os.Exit(1)
	}

	if baseURL == "" {
		baseURL = "https://api.bindist.eu"
	}

	fmt.Printf("Using base URL: %s\n\n", baseURL)

	ctx := context.Background()
	client := bindist.NewClient(baseURL, apiKey)

	// Test: List applications
	fmt.Println("=== List Applications ===")
	apps, err := client.ListApplications(ctx, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !apps.Success {
		fmt.Printf("API Error: %s - %s\n", apps.Error.Code, apps.Error.Message)
		os.Exit(1)
	}

	fmt.Printf("Found %d applications\n", len(apps.Data))
	for _, app := range apps.Data {
		fmt.Printf("  - %s (%s)\n", app.Name, app.ApplicationID)
		if app.Description != "" {
			fmt.Printf("    Description: %s\n", app.Description)
		}
		if len(app.Tags) > 0 {
			fmt.Printf("    Tags: %v\n", app.Tags)
		}
	}
	fmt.Println()

	// If we have applications, test listing versions for the first one
	if len(apps.Data) > 0 {
		appID := apps.Data[0].ApplicationID

		fmt.Printf("=== List Versions for %s ===\n", appID)
		versions, err := client.ListVersions(ctx, appID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if !versions.Success {
			fmt.Printf("API Error: %s - %s\n", versions.Error.Code, versions.Error.Message)
			os.Exit(1)
		}

		fmt.Printf("Found %d versions\n", len(versions.Data))
		for _, ver := range versions.Data {
			fmt.Printf("  - %s (enabled: %v, size: %d bytes)\n", ver.Version, ver.IsEnabled, ver.FileSize)
			if ver.ReleaseNotes != "" {
				fmt.Printf("    Release notes: %s\n", ver.ReleaseNotes)
			}
		}
		fmt.Println()

		// If we have versions, test listing files for the first one
		if len(versions.Data) > 0 {
			version := versions.Data[0].Version

			fmt.Printf("=== List Files for %s v%s ===\n", appID, version)
			files, err := client.ListVersionFiles(ctx, appID, version)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			if !files.Success {
				fmt.Printf("API Error: %s - %s\n", files.Error.Code, files.Error.Message)
				os.Exit(1)
			}

			fmt.Printf("Found %d files\n", len(files.Data))
			for _, f := range files.Data {
				fmt.Printf("  - %s (%s, %d bytes)\n", f.FileName, f.FileType, f.FileSize)
			}
			fmt.Println()

			// Test getting download info
			fmt.Printf("=== Get Download Info for %s v%s ===\n", appID, version)
			download, err := client.GetDownloadInfo(ctx, appID, version, "")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			if !download.Success {
				fmt.Printf("API Error: %s - %s\n", download.Error.Code, download.Error.Message)
				os.Exit(1)
			}

			fmt.Printf("Download URL: %s...\n", download.Data.URL[:50])
			fmt.Printf("File name: %s\n", download.Data.FileName)
			fmt.Printf("File size: %d bytes\n", download.Data.FileSize)
			fmt.Printf("Checksum: %s\n", download.Data.Checksum)
			fmt.Printf("Expires at: %s\n", download.Data.ExpiresAt)
		}
	}

	fmt.Println("\n=== All tests passed! ===")
}
