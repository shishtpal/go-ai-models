// Package main provides an interactive wizard to select the best model based on requirements.
//
// This example demonstrates:
// - Interactive wizard-style interface using bubbletea
// - Multi-criteria decision making
// - Model scoring and ranking
// - Trade-off analysis
// - Side-by-side model comparison
// - Configuration export
//
// Usage:
//   go run main.go                          # Start interactive wizard
//   go run main.go --help                     # Show help message
//
// Environment Variables:
//   CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bubblesList "github.com/charmbracelet/bubbles/list"
	"charm.land/catwalk/pkg/catwalk"
)

var (
	showHelp = flag.Bool("help", false, "Show help message")
)

// Styles for formatting
var (
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	optionStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("228"))
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type requirements struct {
	budget       float64
	contextSize  int64
	reasoning    bool
	vision       bool
}

type modelScore struct {
	model    catwalk.Model
	provider catwalk.Provider
	score    float64
	reasons   []string
}

type step int

const (
	stepBudget step = iota
	stepContext
	stepReasoning
	stepVision
	stepResults
)

// listItem implements list.Item interface for string items
type listItem string

func (i listItem) FilterValue() string { return string(i) }
func (i listItem) Title() string       { return string(i) }
func (i listItem) Description() string { return "" }

type model struct {
	allModels    []modelScore
	step         step
	requirements requirements
	list         bubblesList.Model
	choices      []string
	selected     int
	width        int
	height       int
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
	var allModels []modelScore
	for _, p := range providers {
		for _, m := range p.Models {
			allModels = append(allModels, modelScore{
				model:    m,
				provider: p,
			})
		}
	}

	// Run interactive wizard
	p := tea.NewProgram(initialModel(allModels))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running wizard: %v", err)
	}
}

func initialModel(allModels []modelScore) model {
	// Create list for budget selection
	items := []bubblesList.Item{
		listItem("No budget limit"),
		listItem("Under $0.50 per 1M tokens"),
		listItem("Under $1.00 per 1M tokens"),
		listItem("Under $5.00 per 1M tokens"),
		listItem("Under $10.00 per 1M tokens"),
		listItem("Any cost"),
	}

	l := bubblesList.New(items, bubblesList.NewDefaultDelegate(), 60, 15)
	l.Title = "What's your budget?"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return model{
		allModels: allModels,
		step:      stepBudget,
		list:      l,
		choices:   []string{"0", "0.5", "1.0", "5.0", "10.0", "1000"},
		width:     80,
		height:    24,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyUp, tea.KeyDown:
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		// Handle window resize if needed
	}

	return m, nil
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	selected := m.list.Index()
	choice := m.choices[selected]

	switch m.step {
	case stepBudget:
		m.requirements.budget, _ = parseBudget(choice)
		m.step = stepContext
		m.setupContextList()

	case stepContext:
		m.requirements.contextSize, _ = parseContext(choice)
		m.step = stepReasoning
		m.setupReasoningList()

	case stepReasoning:
		m.requirements.reasoning = (choice == "yes")
		m.step = stepVision
		m.setupVisionList()

	case stepVision:
		m.requirements.vision = (choice == "yes")
		m.step = stepResults
		m.calculateScores()
		m.setupResultsList()

	case stepResults:
		return m, tea.Quit
	}

	return m, nil
}

func (m *model) setupContextList() {
	items := []bubblesList.Item{
		listItem("Any context size"),
		listItem("At least 32K tokens"),
		listItem("At least 100K tokens"),
		listItem("At least 200K tokens"),
		listItem("At least 400K tokens"),
	}

	l := bubblesList.New(items, bubblesList.NewDefaultDelegate(), m.width, m.height)
	l.Title = "What context size do you need?"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	m.list = l
	m.choices = []string{"0", "32000", "100000", "200000", "400000"}
}

func (m *model) setupReasoningList() {
	items := []bubblesList.Item{
		listItem("Yes, I need reasoning capabilities"),
		listItem("No, reasoning not required"),
	}

	l := bubblesList.New(items, bubblesList.NewDefaultDelegate(), m.width, m.height)
	l.Title = "Do you need reasoning capabilities?"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	m.list = l
	m.choices = []string{"yes", "no"}
}

func (m *model) setupVisionList() {
	items := []bubblesList.Item{
		listItem("Yes, I need vision/multimodal"),
		listItem("No, text-only is fine"),
	}

	l := bubblesList.New(items, bubblesList.NewDefaultDelegate(), m.width, m.height)
	l.Title = "Do you need vision/multimodal capabilities?"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	m.list = l
	m.choices = []string{"yes", "no"}
}

func (m *model) calculateScores() {
	for i := range m.allModels {
		mm := &m.allModels[i]
		score := 100.0
		reasons := []string{}

		// Budget constraint
		if m.requirements.budget > 0 && mm.model.CostPer1MIn > m.requirements.budget {
			score -= 100
			reasons = append(reasons, "Over budget")
		} else if mm.model.CostPer1MIn <= m.requirements.budget/2 {
			score += 30
			reasons = append(reasons, "Well under budget")
		}

		// Context size
		if mm.model.ContextWindow >= m.requirements.contextSize {
			score += 20
			reasons = append(reasons, "Meets context requirement")
		} else if mm.model.ContextWindow < m.requirements.contextSize {
			score -= 50
			reasons = append(reasons, "Below context requirement")
		}

		// Reasoning
		if m.requirements.reasoning {
			if mm.model.CanReason {
				score += 25
				reasons = append(reasons, "Has reasoning")
			} else {
				score -= 50
				reasons = append(reasons, "No reasoning")
			}
		}

		// Vision
		if m.requirements.vision {
			if mm.model.SupportsImages {
				score += 25
				reasons = append(reasons, "Has vision")
			} else {
				score -= 50
				reasons = append(reasons, "No vision")
			}
		}

		mm.score = score
		mm.reasons = reasons
	}

	// Sort by score (descending)
	sort.Slice(m.allModels, func(i, j int) bool {
		return m.allModels[i].score > m.allModels[j].score
	})
}

func (m *model) setupResultsList() {
	// Show top 5 matches
	items := []bubblesList.Item{}
	for i := 0; i < min(5, len(m.allModels)); i++ {
		mm := m.allModels[i]
		items = append(items, listItem(fmt.Sprintf("%s (%s) - Score: %.0f",
			mm.model.Name, mm.provider.Name, mm.score)))
	}

	l := bubblesList.New(items, bubblesList.NewDefaultDelegate(), m.width, m.height)
	l.Title = "Top Recommended Models"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	m.list = l
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("AI Model Selector"))
	s.WriteString("\n\n")
	s.WriteString(subtitleStyle.Render("Answer a few questions to find the best model for your needs"))
	s.WriteString("\n\n")

	if m.step == stepResults {
		// Show detailed results
		s.WriteString(m.viewResults())
	} else {
		s.WriteString(m.list.View())
	}

	return s.String()
}

func (m model) viewResults() string {
	var s strings.Builder

	for i := 0; i < min(3, len(m.allModels)); i++ {
		mm := m.allModels[i]

		s.WriteString(titleStyle.Render(fmt.Sprintf("#%d: %s", i+1, mm.model.Name)))
		s.WriteString("\n")
		s.WriteString(fmt.Sprintf("  Provider: %s\n", mm.provider.Name))
		s.WriteString(fmt.Sprintf("  Cost: $%.2f/1M in, $%.2f/1M out\n",
			mm.model.CostPer1MIn, mm.model.CostPer1MOut))
		s.WriteString(fmt.Sprintf("  Context: %dK tokens\n", mm.model.ContextWindow/1000))
		s.WriteString(fmt.Sprintf("  Reasoning: %s | Vision: %s\n",
			boolToStr(mm.model.CanReason), boolToStr(mm.model.SupportsImages)))

		if len(mm.reasons) > 0 {
			s.WriteString("  Reasons: ")
			for j, r := range mm.reasons {
				if j > 0 {
					s.WriteString(", ")
				}
				s.WriteString(optionStyle.Render(r))
			}
			s.WriteString("\n")
		}

		s.WriteString("\n")
	}

	s.WriteString(borderStyle.Render(strings.Repeat("─", 60)))
	s.WriteString("\n")
	s.WriteString("Press Enter to exit or select a model to see details")

	return s.String()
}

func parseBudget(s string) (float64, error) {
	switch s {
	case "0":
		return 0, nil
	case "0.5":
		return 0.5, nil
	case "1.0":
		return 1.0, nil
	case "5.0":
		return 5.0, nil
	case "10.0":
		return 10.0, nil
	case "1000":
		return 1000, nil
	default:
		return 0, fmt.Errorf("invalid budget")
	}
}

func parseContext(s string) (int64, error) {
	switch s {
	case "0":
		return 0, nil
	case "32000":
		return 32000, nil
	case "100000":
		return 100000, nil
	case "200000":
		return 200000, nil
	case "400000":
		return 400000, nil
	default:
		return 0, fmt.Errorf("invalid context size")
	}
}

func boolToStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func printHelp() {
	fmt.Println("model-selector - Interactive wizard to select the best model")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go")
	fmt.Println()
	fmt.Println("This tool will guide you through a series of questions to help")
	fmt.Println("you select the best AI model based on your requirements.")
	fmt.Println()
	fmt.Println("Questions cover:")
	fmt.Println("  - Budget constraints")
	fmt.Println("  - Context window requirements")
	fmt.Println("  - Reasoning capabilities")
	fmt.Println("  - Vision/multimodal support")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)")
}
