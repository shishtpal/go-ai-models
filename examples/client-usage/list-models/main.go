// Package main provides a CLI tool to list models from a specific provider.
//
// This example demonstrates:
// - Fetching models from a specific provider
// - Filtering models by capabilities (reasoning, vision)
// - Sorting models by various criteria
// - Formatting output in table, JSON, and CSV formats
//
// Usage:
//
//	go run main.go --provider openai                    # List all OpenAI models
//	go run main.go --provider anthropic --reasoning       # List reasoning models only
//	go run main.go --provider openai --sort cost          # Sort by cost
//	go run main.go --provider openai --format json        # Output in JSON format
//	go run main.go --provider openai --format csv         # Output in CSV format
//	go run main.go --help                               # Show help message
//
// Environment Variables:
//
//	CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"charm.land/catwalk/pkg/catwalk"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Command-line flags
	providerID   = flag.String("provider", "", "Provider ID (required)")
	reasoning    = flag.Bool("reasoning", false, "Filter by reasoning capability")
	vision       = flag.Bool("vision", false, "Filter by vision capability")
	sortBy       = flag.String("sort", "name", "Sort by: name, cost, context")
	outputFormat = flag.String("format", "table", "Output format: table, json, or csv")
	showHelp     = flag.Bool("help", false, "Show help message")
)

// Styles for table formatting
var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	nameStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	idStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	typeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
	costStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("228"))
	contextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	capStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

func main() {
	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	// Validate required flags
	if *providerID == "" {
		log.Fatal("Error: --provider is required. Use --help for usage information.")
	}

	// Create catwalk client
	client := catwalk.New()
	ctx := context.Background()

	// Fetch providers
	providers, err := client.GetProviders(ctx, "")
	if err != nil {
		if err == catwalk.ErrNotModified {
			log.Println("Data not modified (cached)")
			return
		}
		log.Fatalf("Error fetching providers: %v", err)
	}

	// Find the specified provider
	var provider *catwalk.Provider
	for i := range providers {
		if strings.ToLower(string(providers[i].ID)) == strings.ToLower(*providerID) {
			provider = &providers[i]
			break
		}
	}

	if provider == nil {
		log.Fatalf("Provider not found: %s", *providerID)
	}

	// Filter models
	models := filterModels(provider.Models)

	// Sort models
	sortModels(models, *sortBy)

	// Output in requested format
	switch strings.ToLower(*outputFormat) {
	case "json":
		outputJSON(provider, models)
	case "csv":
		outputCSV(models)
	case "table":
		outputTable(provider, models)
	default:
		log.Fatalf("Unknown format: %s (use 'table', 'json', or 'csv')", *outputFormat)
	}
}

// filterModels applies filters to the model list
func filterModels(models []catwalk.Model) []catwalk.Model {
	var filtered []catwalk.Model

	for _, m := range models {
		// Filter by reasoning capability
		if *reasoning && !m.CanReason {
			continue
		}

		// Filter by vision capability
		if *vision && !m.SupportsImages {
			continue
		}

		filtered = append(filtered, m)
	}

	return filtered
}

// sortModels sorts models by the specified criteria
func sortModels(models []catwalk.Model, sortBy string) {
	switch strings.ToLower(sortBy) {
	case "cost":
		sort.Slice(models, func(i, j int) bool {
			return models[i].CostPer1MIn < models[j].CostPer1MIn
		})
	case "context":
		sort.Slice(models, func(i, j int) bool {
			return models[i].ContextWindow > models[j].ContextWindow
		})
	default: // name
		sort.Slice(models, func(i, j int) bool {
			return models[i].Name < models[j].Name
		})
	}
}

// outputTable displays models in a formatted table
func outputTable(provider *catwalk.Provider, models []catwalk.Model) {
	if len(models) == 0 {
		fmt.Println("No models found matching the criteria.")
		return
	}

	// Print header
	fmt.Printf("%s: %s\n", headerStyle.Render("Provider"), nameStyle.Render(provider.Name))
	fmt.Printf("%s: %s\n", headerStyle.Render("Type"), typeStyle.Render(string(provider.Type)))
	fmt.Printf("%s: %d\n\n", headerStyle.Render("Models"), len(models))

	// Print table header
	fmt.Println(dividerStyle.Render("─┬──────────────────────────────────────────────┬──────────┬─────────┬────────┬────────┐"))
	fmt.Printf("%s %-42s %s %8s %s %7s %s %6s %s %6s %s\n",
		dividerStyle.Render("│"),
		nameStyle.Render("Model Name"),
		dividerStyle.Render("│"),
		costStyle.Render("Cost/1M"),
		dividerStyle.Render("│"),
		contextStyle.Render("Context"),
		dividerStyle.Render("│"),
		capStyle.Render("Reas"),
		dividerStyle.Render("│"),
		capStyle.Render("Vis"),
		dividerStyle.Render("│"))
	fmt.Println(dividerStyle.Render("─┼──────────────────────────────────────────────┼──────────┼─────────┼────────┼────────┤"))

	// Print each model
	for _, m := range models {
		name := m.Name
		if len(name) > 40 {
			name = name[:37] + "..."
		}

		reasoning := " "
		if m.CanReason {
			reasoning = "✓"
		}

		vision := " "
		if m.SupportsImages {
			vision = "✓"
		}

		fmt.Printf("%s %-42s %s %8.2f %s %7dK %s %6s %s %6s %s\n",
			dividerStyle.Render("│"),
			nameStyle.Render(name),
			dividerStyle.Render("│"),
			m.CostPer1MIn,
			dividerStyle.Render("│"),
			m.ContextWindow/1000,
			dividerStyle.Render("│"),
			capStyle.Render(reasoning),
			dividerStyle.Render("│"),
			capStyle.Render(vision),
			dividerStyle.Render("│"))
	}

	fmt.Println(dividerStyle.Render("─┴──────────────────────────────────────────────┴──────────┴─────────┴────────┴────────┘"))
}

// outputJSON displays models in JSON format
func outputJSON(provider *catwalk.Provider, models []catwalk.Model) {
	type ProviderWithModels struct {
		catwalk.Provider
		FilteredModels []catwalk.Model `json:"models"`
	}

	result := ProviderWithModels{
		Provider:       *provider,
		FilteredModels: models,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}
}

// outputCSV displays models in CSV format
func outputCSV(models []catwalk.Model) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Name", "CostPer1MIn", "CostPer1MOut", "ContextWindow", "CanReason", "SupportsImages"}
	if err := writer.Write(header); err != nil {
		log.Fatalf("Error writing CSV header: %v", err)
	}

	// Write rows
	for _, m := range models {
		row := []string{
			m.ID,
			m.Name,
			strconv.FormatFloat(m.CostPer1MIn, 'f', 2, 64),
			strconv.FormatFloat(m.CostPer1MOut, 'f', 2, 64),
			strconv.FormatInt(m.ContextWindow, 10),
			strconv.FormatBool(m.CanReason),
			strconv.FormatBool(m.SupportsImages),
		}
		if err := writer.Write(row); err != nil {
			log.Fatalf("Error writing CSV row: %v", err)
		}
	}
}

// printHelp displays usage information
func printHelp() {
	fmt.Println("list-models - List models from a specific provider")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go --provider <id> [options]")
	fmt.Println()
	fmt.Println("Required Options:")
	fmt.Println("  --provider <id>   Provider ID (e.g., openai, anthropic, google)")
	fmt.Println()
	fmt.Println("Filter Options:")
	fmt.Println("  --reasoning        Filter by reasoning capability")
	fmt.Println("  --vision           Filter by vision capability")
	fmt.Println()
	fmt.Println("Sort Options:")
	fmt.Println("  --sort <field>    Sort by: name (default), cost, context")
	fmt.Println()
	fmt.Println("Output Options:")
	fmt.Println("  --format <fmt>     Output format: table (default), json, csv")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go --provider openai")
	fmt.Println("  go run main.go --provider anthropic --reasoning --sort cost")
	fmt.Println("  go run main.go --provider google --format json")
	fmt.Println("  go run main.go --provider openai --vision --format csv")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)")
}
