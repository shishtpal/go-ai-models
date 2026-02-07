// Package main provides a CLI tool to list all available AI providers.
//
// This example demonstrates:
// - Using the catwalk client to fetch providers
// - Handling ETag support for efficient caching
// - Formatting output in table and JSON formats
// - Filtering providers by type
//
// Usage:
//   go run main.go                    # List all providers in table format
//   go run main.go --type openai       # List only OpenAI-compatible providers
//   go run main.go --format json       # Output in JSON format
//   go run main.go --help             # Show help message
//
// Environment Variables:
//   CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"charm.land/catwalk/pkg/catwalk"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Command-line flags
	providerType = flag.String("type", "", "Filter by provider type (e.g., openai, anthropic, google)")
	outputFormat = flag.String("format", "table", "Output format: table or json")
	showHelp    = flag.Bool("help", false, "Show help message")
)

// Styles for table formatting
var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	nameStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	idStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	typeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	countStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("228"))
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func main() {
	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	// Create catwalk client
	client := catwalk.New()
	ctx := context.Background()

	// Fetch providers with ETag support
	providers, err := client.GetProviders(ctx, "")
	if err != nil {
		if err == catwalk.ErrNotModified {
			log.Println("Data not modified (cached)")
			return
		}
		log.Fatalf("Error fetching providers: %v", err)
	}

	// Filter by provider type if specified
	var filteredProviders []catwalk.Provider
	if *providerType != "" {
		for _, p := range providers {
			if strings.ToLower(string(p.Type)) == strings.ToLower(*providerType) {
				filteredProviders = append(filteredProviders, p)
			}
		}
		providers = filteredProviders
	}

	// Sort providers by name
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Name < providers[j].Name
	})

	// Output in requested format
	switch strings.ToLower(*outputFormat) {
	case "json":
		outputJSON(providers)
	case "table":
		outputTable(providers)
	default:
		log.Fatalf("Unknown format: %s (use 'table' or 'json')", *outputFormat)
	}
}

// outputTable displays providers in a formatted table
func outputTable(providers []catwalk.Provider) {
	if len(providers) == 0 {
		fmt.Println("No providers found.")
		return
	}

	// Print header
	fmt.Println(headerStyle.Render("Available AI Providers"))
	fmt.Println(borderStyle.Render(strings.Repeat("─", 80)))
	fmt.Println()

	// Print each provider
	for _, p := range providers {
		fmt.Printf("%s %s\n", nameStyle.Render(p.Name), idStyle.Render(fmt.Sprintf("(%s)", p.ID)))
		fmt.Printf("  Type: %s\n", typeStyle.Render(string(p.Type)))
		fmt.Printf("  Models: %s\n", countStyle.Render(fmt.Sprintf("%d", len(p.Models))))

		if p.DefaultLargeModelID != "" {
			fmt.Printf("  Default Large: %s\n", p.DefaultLargeModelID)
		}
		if p.DefaultSmallModelID != "" {
			fmt.Printf("  Default Small: %s\n", p.DefaultSmallModelID)
		}

		fmt.Println()
	}

	fmt.Printf(borderStyle.Render("Total: %d providers\n"), len(providers))
}

// outputJSON displays providers in JSON format
func outputJSON(providers []catwalk.Provider) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(providers); err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}
}

// printHelp displays usage information
func printHelp() {
	fmt.Println("list-providers - List all available AI providers")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go                           # List all providers")
	fmt.Println("  go run main.go --type openai               # List OpenAI providers only")
	fmt.Println("  go run main.go --format json               # Output as JSON")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)")
}
