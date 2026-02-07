// Package main provides a CLI tool to display detailed information about a specific model.
//
// This example demonstrates:
// - Finding a specific model across all providers
// - Displaying complete model configuration
// - Showing pricing breakdown (cached/uncached)
// - Displaying reasoning levels and default settings
// - Exporting model configuration as JSON
//
// Usage:
//   go run main.go --model "gpt-4o"                     # Show model info
//   go run main.go --model "claude-3-opus" --provider anthropic  # Specify provider
//   go run main.go --model "gpt-4o" --export              # Export as JSON
//   go run main.go --help                                  # Show help message
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
	"strings"

	"charm.land/catwalk/pkg/catwalk"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Command-line flags
	modelName   = flag.String("model", "", "Model name or ID (required)")
	providerID  = flag.String("provider", "", "Provider ID (optional, if model ID is unique)")
	exportJSON  = flag.Bool("export", false, "Export model configuration as JSON")
	showHelp    = flag.Bool("help", false, "Show help message")
)

// Styles for formatting
var (
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	labelStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("245"))
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	nameStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
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
	if *modelName == "" {
		log.Fatal("Error: --model is required. Use --help for usage information.")
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

	// Find the model
	var foundProvider *catwalk.Provider
	var foundModel *catwalk.Model

	for i := range providers {
		// If provider is specified, skip others
		if *providerID != "" && strings.ToLower(string(providers[i].ID)) != strings.ToLower(*providerID) {
			continue
		}

		for j := range providers[i].Models {
			model := &providers[i].Models[j]
			// Match by ID or name (case-insensitive partial match)
			if strings.EqualFold(model.ID, *modelName) ||
				strings.Contains(strings.ToLower(model.Name), strings.ToLower(*modelName)) {
				foundProvider = &providers[i]
				foundModel = model
				break
			}
		}

		if foundModel != nil {
			break
		}
	}

	if foundModel == nil {
		log.Fatalf("Model not found: %s", *modelName)
	}

	// Export as JSON if requested
	if *exportJSON {
		exportModelJSON(foundProvider, foundModel)
		return
	}

	// Display model information
	displayModelInfo(foundProvider, foundModel)
}

// displayModelInfo shows detailed information about a model
func displayModelInfo(provider *catwalk.Provider, model *catwalk.Model) {
	// Print header
	fmt.Println()
	fmt.Println(headerStyle.Render("Model Information"))
	fmt.Println(borderStyle.Render(strings.Repeat("═", 80)))
	fmt.Println()

	// Basic information
	fmt.Printf("%s %s\n", labelStyle.Render("Name:"), nameStyle.Render(model.Name))
	fmt.Printf("%s %s\n", labelStyle.Render("ID:"), valueStyle.Render(model.ID))
	fmt.Printf("%s %s\n", labelStyle.Render("Provider:"), valueStyle.Render(provider.Name))
	fmt.Printf("%s %s\n", labelStyle.Render("Type:"), valueStyle.Render(string(provider.Type)))
	fmt.Println()

	// Pricing
	fmt.Println(headerStyle.Render("Pricing"))
	fmt.Println(dividerStyle.Render(strings.Repeat("─", 40)))
	fmt.Printf("%s $%.2f per 1M input tokens\n", labelStyle.Render("Input Cost:"), model.CostPer1MIn)
	fmt.Printf("%s $%.2f per 1M output tokens\n", labelStyle.Render("Output Cost:"), model.CostPer1MOut)

	if model.CostPer1MInCached > 0 || model.CostPer1MOutCached > 0 {
		fmt.Println()
		fmt.Println(costStyle.Render("Cached Pricing (with prompt caching):"))
		fmt.Printf("%s $%.2f per 1M cached input tokens\n", labelStyle.Render("Input:"), model.CostPer1MInCached)
		fmt.Printf("%s $%.2f per 1M cached output tokens\n", labelStyle.Render("Output:"), model.CostPer1MOutCached)
	}
	fmt.Println()

	// Capabilities
	fmt.Println(headerStyle.Render("Capabilities"))
	fmt.Println(dividerStyle.Render(strings.Repeat("─", 40)))
	fmt.Printf("%s %dK tokens\n", labelStyle.Render("Context Window:"), model.ContextWindow/1000)
	fmt.Printf("%s %d tokens\n", labelStyle.Render("Default Max Tokens:"), model.DefaultMaxTokens)
	fmt.Printf("%s %s\n", labelStyle.Render("Reasoning:"), capability(model.CanReason))
	fmt.Printf("%s %s\n", labelStyle.Render("Vision:"), capability(model.SupportsImages))
	fmt.Println()

	// Reasoning levels (if applicable)
	if model.CanReason {
		fmt.Println(headerStyle.Render("Reasoning Configuration"))
		fmt.Println(dividerStyle.Render(strings.Repeat("─", 40)))
		if model.DefaultReasoningEffort != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("Default Level:"), valueStyle.Render(model.DefaultReasoningEffort))
		}
		if len(model.ReasoningLevels) > 0 {
			fmt.Printf("%s %s\n", labelStyle.Render("Available Levels:"), valueStyle.Render(strings.Join(model.ReasoningLevels, ", ")))
		}
		fmt.Println()
	}

	// Example usage
	fmt.Println(headerStyle.Render("Example Usage"))
	fmt.Println(dividerStyle.Render(strings.Repeat("─", 40)))
	fmt.Printf("%s\n", labelStyle.Render("Provider Endpoint:"))
	fmt.Printf("  %s\n\n", valueStyle.Render(provider.APIEndpoint))
	fmt.Printf("%s\n", labelStyle.Render("API Key:"))
	fmt.Printf("  %s\n\n", valueStyle.Render(provider.APIKey))
	fmt.Printf("%s\n", labelStyle.Render("Default Headers:"))
	for key, value := range provider.DefaultHeaders {
		fmt.Printf("  %s: %s\n", key, value)
	}
	fmt.Println()

	fmt.Println(borderStyle.Render(strings.Repeat("═", 80)))
}

// capability returns a styled capability indicator
func capability(enabled bool) string {
	if enabled {
		return capStyle.Render("✓ Supported")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("✗ Not supported")
}

// exportModelJSON exports the model configuration as JSON
func exportModelJSON(provider *catwalk.Provider, model *catwalk.Model) {
	type APIConfig struct {
		Endpoint string            `json:"endpoint"`
		APIKey   string            `json:"api_key"`
		Headers  map[string]string `json:"headers"`
	}

	type ModelExport struct {
		Model      catwalk.Model    `json:"model"`
		Provider   catwalk.Provider `json:"provider"`
		APIConfig  APIConfig        `json:"api_config"`
	}

	export := ModelExport{
		Model:    *model,
		Provider: *provider,
		APIConfig: APIConfig{
			Endpoint: provider.APIEndpoint,
			APIKey:   provider.APIKey,
			Headers:  provider.DefaultHeaders,
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(export); err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}
}

// printHelp displays usage information
func printHelp() {
	fmt.Println("model-info - Display detailed information about a specific model")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go --model <name> [options]")
	fmt.Println()
	fmt.Println("Required Options:")
	fmt.Println("  --model <name>    Model name or ID (e.g., gpt-4o, claude-3-opus)")
	fmt.Println()
	fmt.Println("Optional Options:")
	fmt.Println("  --provider <id>    Provider ID (optional, if model ID is unique)")
	fmt.Println("  --export           Export model configuration as JSON")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go --model \"gpt-4o\"")
	fmt.Println("  go run main.go --model \"claude-3-opus\" --provider anthropic")
	fmt.Println("  go run main.go --model \"gpt-4o\" --export > model-config.json")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)")
}
