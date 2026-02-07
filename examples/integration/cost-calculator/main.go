// Package main provides a CLI tool to estimate AI API costs for different models.
//
// This example demonstrates:
// - Using model pricing data from catwalk
// - Calculating costs for input/output token estimates
// - Comparing costs across multiple models
// - Accounting for prompt caching discounts
// - Batch processing multiple scenarios
// - Exporting cost comparisons as CSV/JSON
//
// Usage:
//   go run main.go --model "gpt-4o" --input 1000 --output 500           # Calculate cost
//   go run main.go --compare "gpt-4o,claude-3-opus" --input 1000 --output 500  # Compare models
//   go run main.go --batch scenarios.json --format csv                       # Batch calculation
//   go run main.go --model "gpt-4o" --input 1000 --cached 0.5          # With caching
//   go run main.go --help                                                     # Show help message
//
// Environment Variables:
//   CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)
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
	modelName  = flag.String("model", "", "Model name or ID")
	compareList = flag.String("compare", "", "Comma-separated list of models to compare")
	inputTokens = flag.Int64("input", 0, "Number of input tokens")
	outputTokens = flag.Int64("output", 0, "Number of output tokens")
	cachedRatio = flag.Float64("cached", 0, "Ratio of cached tokens (0-1)")
	batchFile  = flag.String("batch", "", "JSON file with batch scenarios")
	outputFormat = flag.String("format", "table", "Output format: table, json, or csv")
	showHelp   = flag.Bool("help", false, "Show help message")
)

// Styles for formatting
var (
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	modelStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	costStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("228"))
	providerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type costResult struct {
	Model    string  `json:"model"`
	Provider string  `json:"provider"`
	InputCost float64 `json:"input_cost"`
	OutputCost float64 `json:"output_cost"`
	TotalCost float64 `json:"total_cost"`
}

type scenario struct {
	Model       string  `json:"model"`
	InputTokens int64   `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	CachedRatio float64 `json:"cached_ratio"`
}

func main() {
	flag.Parse()

	if *showHelp {
		printHelp()
		return
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

	// Handle batch mode
	if *batchFile != "" {
		processBatch(providers, *batchFile)
		return
	}

	// Handle compare mode
	if *compareList != "" {
		compareModels(providers, strings.Split(*compareList, ","))
		return
	}

	// Single model calculation
	if *modelName == "" {
		log.Fatal("Error: --model is required. Use --help for usage information.")
	}

	if *inputTokens == 0 || *outputTokens == 0 {
		log.Fatal("Error: --input and --output are required.")
	}

	result := calculateCost(providers, *modelName, *inputTokens, *outputTokens, *cachedRatio)
	if result == nil {
		log.Fatalf("Model not found: %s", *modelName)
	}

	displayCostResult([]costResult{*result})
}

// calculateCost calculates cost for a single model
func calculateCost(providers []catwalk.Provider, modelName string, inputTokens, outputTokens int64, cachedRatio float64) *costResult {
	var model *catwalk.Model
	var provider *catwalk.Provider

	// Find model
	for i := range providers {
		for j := range providers[i].Models {
			if strings.EqualFold(providers[i].Models[j].ID, modelName) ||
				strings.Contains(strings.ToLower(providers[i].Models[j].Name), strings.ToLower(modelName)) {
				model = &providers[i].Models[j]
				provider = &providers[i]
				break
			}
		}
		if model != nil {
			break
		}
	}

	if model == nil {
		return nil
	}

	// Calculate costs
	cachedInputTokens := float64(inputTokens) * cachedRatio
	uncachedInputTokens := float64(inputTokens) * (1 - cachedRatio)

	inputCost := (uncachedInputTokens * model.CostPer1MIn / 1_000_000) +
		(cachedInputTokens * model.CostPer1MInCached / 1_000_000)

	outputCost := float64(outputTokens) * model.CostPer1MOut / 1_000_000

	return &costResult{
		Model:     model.Name,
		Provider:  provider.Name,
		InputCost:  inputCost,
		OutputCost: outputCost,
		TotalCost: inputCost + outputCost,
	}
}

// compareModels compares costs across multiple models
func compareModels(providers []catwalk.Provider, modelNames []string) {
	var results []costResult

	for _, name := range modelNames {
		name = strings.TrimSpace(name)
		result := calculateCost(providers, name, *inputTokens, *outputTokens, *cachedRatio)
		if result != nil {
			results = append(results, *result)
		}
	}

	if len(results) == 0 {
		fmt.Println("No models found.")
		return
	}

	// Sort by total cost
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalCost < results[j].TotalCost
	})

	displayCostResult(results)
}

// processBatch processes multiple scenarios from a file
func processBatch(providers []catwalk.Provider, batchFile string) {
	// Read batch file
	data, err := os.ReadFile(batchFile)
	if err != nil {
		log.Fatalf("Error reading batch file: %v", err)
	}

	var scenarios []scenario
	if err := json.Unmarshal(data, &scenarios); err != nil {
		log.Fatalf("Error parsing batch file: %v", err)
	}

	var results []costResult
	for _, s := range scenarios {
		result := calculateCost(providers, s.Model, s.InputTokens, s.OutputTokens, s.CachedRatio)
		if result != nil {
			results = append(results, *result)
		}
	}

	if len(results) == 0 {
		fmt.Println("No valid scenarios found.")
		return
	}

	displayCostResult(results)
}

// displayCostResult displays cost results
func displayCostResult(results []costResult) {
	switch strings.ToLower(*outputFormat) {
	case "json":
		outputJSON(results)
	case "csv":
		outputCSV(results)
	case "table":
		outputTable(results)
	default:
		log.Fatalf("Unknown format: %s (use 'table', 'json', or 'csv')", *outputFormat)
	}
}

// outputTable displays results in a formatted table
func outputTable(results []costResult) {
	if len(results) == 0 {
		fmt.Println("No results to display.")
		return
	}

	fmt.Println()
	fmt.Println(headerStyle.Render("Cost Calculation Results"))
	fmt.Println(borderStyle.Render(strings.Repeat("═", 80)))
	fmt.Println()

	fmt.Println(dividerStyle.Render("─┬──────────────────────────────────────────────┬──────────┬─────────┬────────┐"))
	fmt.Printf("%s %-42s %s %8s %s %7s %s %6s %s\n",
		dividerStyle.Render("│"),
		modelStyle.Render("Model"),
		dividerStyle.Render("│"),
		costStyle.Render("Input"),
		dividerStyle.Render("│"),
		costStyle.Render("Output"),
		dividerStyle.Render("│"),
		costStyle.Render("Total"),
		dividerStyle.Render("│"))
	fmt.Println(dividerStyle.Render("─┼──────────────────────────────────────────────┼──────────┼─────────┼────────┤"))

	for _, r := range results {
		name := r.Model
		if len(name) > 40 {
			name = name[:37] + "..."
		}

		fmt.Printf("%s %-42s %s $%7.4f %s $%7.4f %s $%6.4f %s\n",
			dividerStyle.Render("│"),
			name,
			dividerStyle.Render("│"),
			r.InputCost,
			dividerStyle.Render("│"),
			r.OutputCost,
			dividerStyle.Render("│"),
			r.TotalCost,
			dividerStyle.Render("│"))
	}

	fmt.Println(dividerStyle.Render("─┴──────────────────────────────────────────────┴──────────┴─────────┴────────┘"))

	// Show provider information
	fmt.Println()
	fmt.Println(headerStyle.Render("Provider Information"))
	for _, r := range results {
		fmt.Printf("%s: %s\n", modelStyle.Render(r.Model), providerStyle.Render(r.Provider))
	}
}

// outputJSON displays results in JSON format
func outputJSON(results []costResult) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}
}

// outputCSV displays results in CSV format
func outputCSV(results []costResult) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	header := []string{"Model", "Provider", "InputCost", "OutputCost", "TotalCost"}
	if err := writer.Write(header); err != nil {
		log.Fatalf("Error writing CSV header: %v", err)
	}

	// Write rows
	for _, r := range results {
		row := []string{
			r.Model,
			r.Provider,
			strconv.FormatFloat(r.InputCost, 'f', 4, 64),
			strconv.FormatFloat(r.OutputCost, 'f', 4, 64),
			strconv.FormatFloat(r.TotalCost, 'f', 4, 64),
		}
		if err := writer.Write(row); err != nil {
			log.Fatalf("Error writing CSV row: %v", err)
		}
	}
}

// printHelp displays usage information
func printHelp() {
	fmt.Println("cost-calculator - Estimate AI API costs for different models")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go --model <name> --input <tokens> --output <tokens> [options]")
	fmt.Println()
	fmt.Println("Required Options:")
	fmt.Println("  --model <name>     Model name or ID")
	fmt.Println("  --input <tokens>    Number of input tokens")
	fmt.Println("  --output <tokens>   Number of output tokens")
	fmt.Println()
	fmt.Println("Optional Options:")
	fmt.Println("  --cached <ratio>    Ratio of cached tokens (0-1, default: 0)")
	fmt.Println("  --compare <models>  Comma-separated list of models to compare")
	fmt.Println("  --batch <file>      JSON file with batch scenarios")
	fmt.Println("  --format <fmt>      Output format: table (default), json, csv")
	fmt.Println()
	fmt.Println("Batch File Format (JSON):")
	fmt.Println("  [")
	fmt.Println("    {")
	fmt.Println("      \"model\": \"gpt-4o\",")
	fmt.Println("      \"input_tokens\": 1000,")
	fmt.Println("      \"output_tokens\": 500,")
	fmt.Println("      \"cached_ratio\": 0.5")
	fmt.Println("    },")
	fmt.Println("    ...")
	fmt.Println("  ]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go --model \"gpt-4o\" --input 1000 --output 500")
	fmt.Println("  go run main.go --compare \"gpt-4o,claude-3-opus\" --input 1000 --output 500")
	fmt.Println("  go run main.go --model \"gpt-4o\" --input 1000 --output 500 --cached 0.5")
	fmt.Println("  go run main.go --batch scenarios.json --format csv")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)")
}
