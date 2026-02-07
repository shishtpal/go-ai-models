// Package main provides an interactive CLI chat bot that uses catwalk to select models
// and make actual API calls to AI providers.
//
// This example demonstrates:
// - Integrating catwalk with AI API calls using OpenAI-compatible client
// - Real-time cost estimation using catwalk pricing data
// - Model selection based on use case
// - Interactive CLI chat interface
// - Handling different provider types (openai, openai-compat, anthropic, etc.)
// - Conversation history management
//
// Usage:
//
//	go run main.go --provider openai --model gpt-4o           # Start with specific model
//	go run main.go --provider anthropic                       # Use default model
//	go run main.go --provider openai --system "You are a helpful coding assistant"
//	go run main.go --help                                     # Show help message
//
// Environment Variables:
//
//	CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"charm.land/catwalk/pkg/catwalk"
	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"
)

var (
	providerID   = flag.String("provider", "", "Provider ID (e.g., openai, anthropic)")
	modelName    = flag.String("model", "", "Model ID (overrides default)")
	systemPrompt = flag.String("system", "", "System prompt for the conversation")
	maxTokens    = flag.Int("max-tokens", 0, "Max tokens for response (0 = model default)")
	apiKey       = flag.String("api-key", "", "API key (overrides provider config)")
	debug        = flag.Bool("debug", false, "Show debug information")
	showHelp     = flag.Bool("help", false, "Show help message")
)

// Styles for formatting
var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	userStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	aiStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("120"))
	costStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("228"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	promptStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
)

type chatSession struct {
	client      *openai.Client
	provider    *catwalk.Provider
	model       *catwalk.Model
	messages    []openai.ChatCompletionMessage
	totalTokens int
	totalCost   float64
}

func main() {
	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	if *providerID == "" {
		log.Fatal("Error: --provider is required. Use --help for usage information.")
	}

	// Create catwalk client and fetch providers
	catwalkClient := catwalk.New()
	ctx := context.Background()

	providers, err := catwalkClient.GetProviders(ctx, "")
	if err != nil {
		log.Fatalf("Error fetching providers: %v", err)
	}

	// Find provider
	var provider *catwalk.Provider
	for i := range providers {
		if strings.EqualFold(string(providers[i].ID), *providerID) {
			provider = &providers[i]
			break
		}
	}

	if provider == nil {
		fmt.Println(errorStyle.Render("Provider not found: " + *providerID))
		fmt.Println(infoStyle.Render("\nAvailable providers:"))
		for _, p := range providers {
			fmt.Printf("  - %s (%s)\n", p.ID, p.Name)
		}
		os.Exit(1)
	}

	// Find model
	var model *catwalk.Model
	if *modelName != "" {
		for i := range provider.Models {
			if strings.EqualFold(provider.Models[i].ID, *modelName) {
				model = &provider.Models[i]
				break
			}
		}
		if model == nil {
			fmt.Println(errorStyle.Render("Model not found: " + *modelName))
			fmt.Println(infoStyle.Render("\nAvailable models for " + provider.Name + ":"))
			for _, m := range provider.Models {
				fmt.Printf("  - %s (%s)\n", m.ID, m.Name)
			}
			os.Exit(1)
		}
	} else {
		// Use default model
		modelID := provider.DefaultLargeModelID
		for i := range provider.Models {
			if provider.Models[i].ID == modelID {
				model = &provider.Models[i]
				break
			}
		}
		if model == nil && len(provider.Models) > 0 {
			model = &provider.Models[0]
		}
	}

	if model == nil {
		log.Fatal("No model found for provider.")
	}

	// Resolve API key (flag > env var > provider config)
	resolvedAPIKey := resolveAPIKey(provider)
	if resolvedAPIKey == "" {
		fmt.Println(errorStyle.Render("No API key found!"))
		fmt.Println(infoStyle.Render("\nProvide an API key via:"))
		fmt.Println("  --api-key <key>")
		fmt.Printf("  %s environment variable\n", getEnvKeyName(provider.ID))
		os.Exit(1)
	}

	// Create OpenAI-compatible client
	client := createClient(provider, resolvedAPIKey)

	// Debug info
	if *debug {
		fmt.Println(infoStyle.Render("\n[Debug Info]"))
		fmt.Printf("  Endpoint: %s\n", provider.APIEndpoint)
		fmt.Printf("  API Key: %s...%s\n", resolvedAPIKey[:4], resolvedAPIKey[len(resolvedAPIKey)-4:])
		fmt.Printf("  Type: %s\n", provider.Type)
		if len(provider.DefaultHeaders) > 0 {
			fmt.Println("  Headers:")
			for k, v := range provider.DefaultHeaders {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
		fmt.Println()
	}

	// Create chat session
	session := &chatSession{
		client:   client,
		provider: provider,
		model:    model,
		messages: []openai.ChatCompletionMessage{},
	}

	// Add system prompt if provided
	if *systemPrompt != "" {
		session.messages = append(session.messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: *systemPrompt,
		})
	}

	// Print header
	printHeader(provider, model)

	// Start chat loop
	runChatLoop(session)
}

func resolveAPIKey(provider *catwalk.Provider) string {
	// Priority: flag > env var > provider config
	if *apiKey != "" {
		return *apiKey
	}

	// Check environment variable based on provider
	envKey := getEnvKeyName(provider.ID)
	if key := os.Getenv(envKey); key != "" {
		return key
	}

	// Fall back to provider config
	return provider.APIKey
}

func getEnvKeyName(providerID catwalk.InferenceProvider) string {
	switch providerID {
	case catwalk.InferenceProviderOpenAI:
		return "OPENAI_API_KEY"
	case catwalk.InferenceProviderAnthropic:
		return "ANTHROPIC_API_KEY"
	case catwalk.InferenceProviderGemini:
		return "GEMINI_API_KEY"
	case catwalk.InferenceProviderXAI:
		return "XAI_API_KEY"
	case catwalk.InferenceProviderGROQ:
		return "GROQ_API_KEY"
	case catwalk.InferenceProviderOpenRouter:
		return "OPENROUTER_API_KEY"
	case catwalk.InferenceProviderCerebras:
		return "CEREBRAS_API_KEY"
	case catwalk.InferenceProviderHuggingFace:
		return "HUGGINGFACE_API_KEY"
	case catwalk.InferenceProviderVercel:
		return "VERCEL_API_KEY"
	default:
		return strings.ToUpper(string(providerID)) + "_API_KEY"
	}
}

// headerTransport adds custom headers to all requests
type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.base.RoundTrip(req)
}

func createClient(provider *catwalk.Provider, apiKey string) *openai.Client {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = provider.APIEndpoint

	// Add custom headers if provider has them
	if len(provider.DefaultHeaders) > 0 {
		transport := &headerTransport{
			base:    http.DefaultTransport,
			headers: provider.DefaultHeaders,
		}
		config.HTTPClient = &http.Client{Transport: transport}
	}

	return openai.NewClientWithConfig(config)
}

func printHeader(provider *catwalk.Provider, model *catwalk.Model) {
	fmt.Println()
	fmt.Println(headerStyle.Render("AI Chat Bot"))
	fmt.Println(borderStyle.Render(strings.Repeat("─", 60)))
	fmt.Println()
	fmt.Printf("%s %s\n", infoStyle.Render("Provider:"), provider.Name)
	fmt.Printf("%s %s\n", infoStyle.Render("Model:"), model.Name)
	fmt.Printf("%s $%.4f/1M input, $%.4f/1M output\n",
		infoStyle.Render("Pricing:"),
		model.CostPer1MIn,
		model.CostPer1MOut)
	fmt.Printf("%s %dK tokens\n", infoStyle.Render("Context:"), model.ContextWindow/1000)
	fmt.Println()
	fmt.Println(borderStyle.Render(strings.Repeat("─", 60)))
	fmt.Println(infoStyle.Render("Type your message and press Enter. Commands:"))
	fmt.Println(infoStyle.Render("  /clear  - Clear conversation history"))
	fmt.Println(infoStyle.Render("  /cost   - Show current session cost"))
	fmt.Println(infoStyle.Render("  /quit   - Exit the chat"))
	fmt.Println(borderStyle.Render(strings.Repeat("─", 60)))
	fmt.Println()
}

func runChatLoop(session *chatSession) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Print prompt
		fmt.Print(promptStyle.Render("You: "))

		// Read input
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nGoodbye!")
				return
			}
			log.Fatalf("Error reading input: %v", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			if handleCommand(session, input) {
				continue
			} else {
				return // /quit command
			}
		}

		// Add user message
		session.messages = append(session.messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})

		// Make API call
		fmt.Print(aiStyle.Render("AI: "))

		response, err := sendMessage(session)
		if err != nil {
			fmt.Println()
			fmt.Println(errorStyle.Render("Error: " + err.Error()))
			// Remove the failed user message
			session.messages = session.messages[:len(session.messages)-1]
			continue
		}

		// Print response
		fmt.Println(response.content)

		// Add assistant message to history
		session.messages = append(session.messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: response.content,
		})

		// Update and show cost
		session.totalTokens += response.inputTokens + response.outputTokens
		session.totalCost += response.cost

		fmt.Printf("%s tokens: %d (in: %d, out: %d) | cost: $%.6f | session: $%.6f\n",
			costStyle.Render("→"),
			response.inputTokens+response.outputTokens,
			response.inputTokens,
			response.outputTokens,
			response.cost,
			session.totalCost)
		fmt.Println()
	}
}

func handleCommand(session *chatSession, cmd string) bool {
	switch strings.ToLower(cmd) {
	case "/quit", "/exit", "/q":
		fmt.Println()
		fmt.Println(infoStyle.Render("Session Summary:"))
		fmt.Printf("  Total tokens: %d\n", session.totalTokens)
		fmt.Printf("  Total cost: $%.6f\n", session.totalCost)
		fmt.Println()
		fmt.Println("Goodbye!")
		return false

	case "/clear":
		// Keep system message if present
		var systemMsg []openai.ChatCompletionMessage
		if len(session.messages) > 0 && session.messages[0].Role == openai.ChatMessageRoleSystem {
			systemMsg = session.messages[:1]
		}
		session.messages = systemMsg
		fmt.Println(infoStyle.Render("Conversation cleared."))
		fmt.Println()
		return true

	case "/cost":
		fmt.Println()
		fmt.Println(infoStyle.Render("Session Statistics:"))
		fmt.Printf("  Messages: %d\n", len(session.messages))
		fmt.Printf("  Total tokens: %d\n", session.totalTokens)
		fmt.Printf("  Total cost: $%.6f\n", session.totalCost)
		fmt.Println()
		return true

	case "/help":
		fmt.Println()
		fmt.Println(infoStyle.Render("Available commands:"))
		fmt.Println("  /clear  - Clear conversation history")
		fmt.Println("  /cost   - Show current session cost")
		fmt.Println("  /help   - Show this help")
		fmt.Println("  /quit   - Exit the chat")
		fmt.Println()
		return true

	default:
		fmt.Println(errorStyle.Render("Unknown command: " + cmd))
		fmt.Println(infoStyle.Render("Type /help for available commands."))
		fmt.Println()
		return true
	}
}

type apiResponse struct {
	content      string
	inputTokens  int
	outputTokens int
	cost         float64
}

func sendMessage(session *chatSession) (*apiResponse, error) {
	ctx := context.Background()

	// Build request
	req := openai.ChatCompletionRequest{
		Model:    session.model.ID,
		Messages: session.messages,
	}

	// Set max tokens if specified
	if *maxTokens > 0 {
		req.MaxTokens = *maxTokens
	} else if session.model.DefaultMaxTokens > 0 {
		req.MaxTokens = int(session.model.DefaultMaxTokens)
	}

	// Make API call
	resp, err := session.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from model")
	}

	// Calculate cost
	inputTokens := resp.Usage.PromptTokens
	outputTokens := resp.Usage.CompletionTokens
	cost := calculateCost(session.model, inputTokens, outputTokens)

	return &apiResponse{
		content:      resp.Choices[0].Message.Content,
		inputTokens:  inputTokens,
		outputTokens: outputTokens,
		cost:         cost,
	}, nil
}

func calculateCost(model *catwalk.Model, inputTokens, outputTokens int) float64 {
	inputCost := float64(inputTokens) * model.CostPer1MIn / 1_000_000
	outputCost := float64(outputTokens) * model.CostPer1MOut / 1_000_000
	return inputCost + outputCost
}

func printHelp() {
	fmt.Println("chat-bot - Interactive CLI chat bot with catwalk integration")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go --provider <id> [options]")
	fmt.Println()
	fmt.Println("Required:")
	fmt.Println("  --provider <id>     Provider ID (e.g., openai, anthropic, google)")
	fmt.Println()
	fmt.Println("Optional:")
	fmt.Println("  --model <id>        Model ID (uses provider default if not specified)")
	fmt.Println("  --system <prompt>   System prompt for the conversation")
	fmt.Println("  --max-tokens <n>    Max tokens for response (0 = model default)")
	fmt.Println("  --api-key <key>     API key (overrides env var and provider config)")
	fmt.Println("  --debug             Show debug information (endpoint, headers, etc.)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go --provider openai --model gpt-4o")
	fmt.Println("  go run main.go --provider anthropic")
	fmt.Println("  go run main.go --provider openai --system \"You are a helpful coding assistant\"")
	fmt.Println("  go run main.go --provider openai --api-key sk-xxx --debug")
	fmt.Println()
	fmt.Println("In-chat commands:")
	fmt.Println("  /clear   Clear conversation history")
	fmt.Println("  /cost    Show current session cost")
	fmt.Println("  /help    Show available commands")
	fmt.Println("  /quit    Exit the chat")
	fmt.Println()
	fmt.Println("Environment Variables (checked if --api-key not provided):")
	fmt.Println("  OPENAI_API_KEY      - for OpenAI provider")
	fmt.Println("  ANTHROPIC_API_KEY   - for Anthropic provider")
	fmt.Println("  GEMINI_API_KEY      - for Google/Gemini provider")
	fmt.Println("  GROQ_API_KEY        - for Groq provider")
	fmt.Println("  OPENROUTER_API_KEY  - for OpenRouter provider")
	fmt.Println("  (or <PROVIDER>_API_KEY for others)")
	fmt.Println()
	fmt.Println("  CATWALK_URL - URL of the catwalk service (default: http://localhost:8080)")
}
