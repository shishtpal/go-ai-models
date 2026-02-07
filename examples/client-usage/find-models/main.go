// Package main provides a CLI tool to find models matching specific criteria.
//
// This example demonstrates:
// - Searching models across all providers
// - Filtering by multiple criteria (cost, context, reasoning, vision)
// - Interactive mode for step-by-step filtering using bubbletea
// - Scoring and ranking models
// - Side-by-side model comparison
//
// Usage:
//   go run main.go --max-cost 1.0 --min-context 100000       # Non-interactive search
//   go run main.go --reasoning --vision                         # Filter by capabilities
//   go run main.go --interactive                                # Interactive mode
//   go run main.go --compare "gpt-4o,claude-3-opus"          # Compare specific models
//   go run main.go --help                                      # Show help message
//
// Environment Variables:
//   CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"charm.land/catwalk/pkg/catwalk"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Command-line flags (for non-interactive mode)
	maxCost       = flag.Float64("max-cost", 0, "Maximum cost per 1M input tokens (0 = no limit)")
	minContext    = flag.Int64("min-context", 0, "Minimum context window (0 = no limit)")
	reasoning     = flag.Bool("reasoning", false, "Filter by reasoning capability")
	vision        = flag.Bool("vision", false, "Filter by vision capability")
	interactive   = flag.Bool("interactive", false, "Interactive mode")
	compareModels = flag.String("compare", "", "Comma-separated list of models to compare")
	showHelp      = flag.Bool("help", false, "Show help message")
)

// Styles for formatting
var (
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	nameStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	scoreStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("228"))
	costStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("228"))
	contextStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	providerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type modelMatch struct {
	model    catwalk.Model
	provider catwalk.Provider
	score    float64
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

	// Collect all models
	var allModels []modelMatch
	for _, p := range providers {
		for _, m := range p.Models {
			allModels = append(allModels, modelMatch{
				model:    m,
				provider: p,
			})
		}
	}

	// Handle different modes
	if *compareModels != "" {
		compareModelsList(providers, strings.Split(*compareModels, ","))
		return
	}

	if *interactive {
		runInteractiveMode(allModels)
		return
	}

	// Non-interactive search
	matches := filterModels(allModels, *maxCost, *minContext, *reasoning, *vision)
	if len(matches) == 0 {
		fmt.Println("No models found matching criteria.")
		return
	}

	displayMatches(matches)
}

// filterModels applies filters to model list
func filterModels(models []modelMatch, maxCost float64, minContext int64, reasoning, vision bool) []modelMatch {
	var filtered []modelMatch

	for _, mm := range models {
		// Filter by max cost
		if maxCost > 0 && mm.model.CostPer1MIn > maxCost {
			continue
		}

		// Filter by min context
		if minContext > 0 && mm.model.ContextWindow < minContext {
			continue
		}

		// Filter by reasoning capability
		if reasoning && !mm.model.CanReason {
			continue
		}

		// Filter by vision capability
		if vision && !mm.model.SupportsImages {
			continue
		}

		filtered = append(filtered, mm)
	}

	return filtered
}

// scoreModels calculates match scores for models
func scoreModels(models []modelMatch) []modelMatch {
	for i := range models {
		mm := &models[i]

		// Base score
		score := 100.0

		// Penalize high cost (lower cost = higher score)
		if mm.model.CostPer1MIn > 0 {
			costPenalty := math.Min(mm.model.CostPer1MIn/10.0, 50.0)
			score -= costPenalty
		}

		// Bonus for large context window
		if mm.model.ContextWindow >= 200000 {
			score += 20
		} else if mm.model.ContextWindow >= 100000 {
			score += 10
		}

		// Bonus for reasoning capability
		if mm.model.CanReason {
			score += 15
		}

		// Bonus for vision capability
		if mm.model.SupportsImages {
			score += 10
		}

		mm.score = score
	}

	// Sort by score (descending)
	sort.Slice(models, func(i, j int) bool {
		return models[i].score > models[j].score
	})

	return models
}

// displayMatches shows matching models
func displayMatches(models []modelMatch) {
	// Score models
	models = scoreModels(models)

	fmt.Println()
	fmt.Println(headerStyle.Render("Matching Models"))
	fmt.Println(borderStyle.Render(strings.Repeat("═", 80)))
	fmt.Println()

	for i, mm := range models {
		if i >= 10 { // Show top 10 matches
			break
		}

		fmt.Printf("%s #%d %s\n",
			scoreStyle.Render(fmt.Sprintf("[%.0f]", mm.score)),
			i+1,
			nameStyle.Render(mm.model.Name))
		fmt.Printf("  Provider: %s\n", providerStyle.Render(mm.provider.Name))
		fmt.Printf("  Cost: $%.2f/1M in, $%.2f/1M out | Context: %dK\n",
			mm.model.CostPer1MIn, mm.model.CostPer1MOut, mm.model.ContextWindow/1000)

		if mm.model.CanReason {
			fmt.Printf("  %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Render("✓ Reasoning"))
		}
		if mm.model.SupportsImages {
			fmt.Printf("  %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Render("✓ Vision"))
		}

		fmt.Println()
	}

	fmt.Printf(borderStyle.Render("Showing top %d of %d matches\n"), min(10, len(models)), len(models))
}

// compareModelsList compares specific models side-by-side
func compareModelsList(providers []catwalk.Provider, modelNames []string) {
	var models []struct {
		model    catwalk.Model
		provider catwalk.Provider
	}

	// Find models
	for _, name := range modelNames {
		name = strings.TrimSpace(name)
		for _, p := range providers {
			for _, m := range p.Models {
				if strings.EqualFold(m.ID, name) ||
					strings.Contains(strings.ToLower(m.Name), strings.ToLower(name)) {
					models = append(models, struct {
						model    catwalk.Model
						provider catwalk.Provider
					}{m, p})
					break
				}
			}
		}
	}

	if len(models) == 0 {
		fmt.Println("No models found.")
		return
	}

	// Display comparison
	fmt.Println()
	fmt.Println(headerStyle.Render("Model Comparison"))
	fmt.Println(borderStyle.Render(strings.Repeat("═", 80)))
	fmt.Println()

	for _, m := range models {
		fmt.Printf("%s\n", nameStyle.Render(m.model.Name))
		fmt.Printf("  Provider: %s\n", providerStyle.Render(m.provider.Name))
		fmt.Printf("  Cost: $%.2f/1M in, $%.2f/1M out\n",
			m.model.CostPer1MIn, m.model.CostPer1MOut)
		fmt.Printf("  Context: %dK tokens\n", m.model.ContextWindow/1000)
		fmt.Printf("  Reasoning: %s | Vision: %s\n",
			boolToStr(m.model.CanReason), boolToStr(m.model.SupportsImages))
		fmt.Println()
	}
}

// runInteractiveMode runs interactive filtering interface
func runInteractiveMode(models []modelMatch) {
	p := tea.NewProgram(initialModel(models))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running interactive mode: %v", err)
	}
}

// initialModel creates initial model for interactive interface
func initialModel(models []modelMatch) model {
	return model{
		models:      models,
		filtered:    models,
		step:         stepMaxCost,
		currentInput: "",
	}
}

// Model for interactive interface
type model struct {
	models       []modelMatch
	filtered     []modelMatch
	step         step
	currentInput string
}

type step int

const (
	stepMaxCost step = iota
	stepMinContext
	stepCapabilities
	stepResults
)

// Init initializes model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			switch m.step {
			case stepMaxCost:
				if cost, err := strconv.ParseFloat(m.currentInput, 64); err == nil {
					m.filtered = filterModels(m.models, cost, 0, false, false)
					m.step = stepMinContext
					m.currentInput = ""
				}
			case stepMinContext:
				if ctx, err := strconv.ParseInt(m.currentInput, 10, 64); err == nil {
					m.filtered = filterModels(m.filtered, 0, ctx, false, false)
					m.step = stepCapabilities
				}
			case stepCapabilities:
				m.step = stepResults
			case stepResults:
				return m, tea.Quit
			}

		case tea.KeyBackspace:
			if len(m.currentInput) > 0 {
				m.currentInput = m.currentInput[:len(m.currentInput)-1]
			}

		default:
			// Handle character input for numeric values
			if (m.step == stepMaxCost || m.step == stepMinContext) && len(msg.String()) == 1 {
				char := msg.String()
				// Only allow digits and decimal point
				if (char >= "0" && char <= "9") || char == "." {
					m.currentInput += char
				}
			}
		}
	}

	return m, nil
}

// View renders interface
func (m model) View() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("Find Models - Interactive Mode"))
	s.WriteString("\n\n")

	switch m.step {
	case stepMaxCost:
		s.WriteString("Enter maximum cost per 1M input tokens (or press Enter to skip): ")
		s.WriteString(m.currentInput)
	case stepMinContext:
		s.WriteString(fmt.Sprintf("Filtered to %d models\n\n", len(m.filtered)))
		s.WriteString("Enter minimum context window in K (or press Enter to skip): ")
		s.WriteString(m.currentInput)
	case stepCapabilities:
		s.WriteString(fmt.Sprintf("Filtered to %d models\n\n", len(m.filtered)))
		s.WriteString("Press Enter to continue to results...")
	case stepResults:
		m.filtered = scoreModels(m.filtered)
		s.WriteString(fmt.Sprintf("Found %d matching models\n\n", len(m.filtered)))
		for i, mm := range m.filtered {
			if i >= 5 {
				break
			}
			s.WriteString(fmt.Sprintf("%d. %s (%s) - $%.2f/1M in\n",
				i+1, mm.model.Name, mm.provider.Name, mm.model.CostPer1MIn))
		}
		s.WriteString("\nPress Enter to exit...")
	}

	return s.String()
}

// boolToStr converts boolean to string
func boolToStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// printHelp displays usage information
func printHelp() {
	fmt.Println("find-models - Find models matching specific criteria")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go [options]")
	fmt.Println()
	fmt.Println("Filter Options:")
	fmt.Println("  --max-cost <float>      Maximum cost per 1M input tokens (0 = no limit)")
	fmt.Println("  --min-context <int>     Minimum context window (0 = no limit)")
	fmt.Println("  --reasoning              Filter by reasoning capability")
	fmt.Println("  --vision                Filter by vision capability")
	fmt.Println()
	fmt.Println("Interactive Options:")
	fmt.Println("  --interactive            Interactive filtering mode")
	fmt.Println("  --compare <models>      Comma-separated list of models to compare")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go --max-cost 1.0 --min-context 100000")
	fmt.Println("  go run main.go --reasoning --vision")
	fmt.Println("  go run main.go --interactive")
	fmt.Println("  go run main.go --compare \"gpt-4o,claude-3-opus\"")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
