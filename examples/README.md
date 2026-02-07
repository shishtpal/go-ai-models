# Catwalk Examples

This directory contains practical CLI tool examples demonstrating how to use the catwalk client library and integrate it into real applications.

## Overview

The examples are organized into two categories:

- **Client Usage Examples** - Simple tools demonstrating catwalk client API usage
- **Integration Examples** - More complex examples showing real-world integration patterns

## Prerequisites

1. **Catwalk Server Running**: Most examples require the catwalk server to be running.
   ```bash
   # From the project root
   go run main.go
   ```

2. **Environment Variables**: Set the `CATWALK_URL` if your server is not at localhost:8080:
   ```bash
   export CATWALK_URL=http://localhost:8080
   # Windows PowerShell
   $env:CATWALK_URL = "http://localhost:8080"
   ```

3. **API Keys**: Some integration examples may require provider API keys:
   ```bash
   export OPENAI_API_KEY=your-key-here
   export ANTHROPIC_API_KEY=your-key-here
   export GEMINI_API_KEY=your-key-here
   ```

## Quick Start

### 1. List All Providers

```bash
cd examples/client-usage/list-providers
go run main.go
```

Output:
```
Available AI Providers
────────────────────────────────────────────────────────────────────────────────

OpenAI (openai)
  Type: openai
  Models: 12
  Default Large: gpt-5.1-codex
  Default Small: gpt-4o

Anthropic (anthropic)
  Type: anthropic
  Models: 10
  Default Large: claude-sonnet-4-5-20250929
  Default Small: claude-3-5-haiku-20241022

...
```

### 2. List Models from a Provider

```bash
cd examples/client-usage/list-models
go run main.go --provider openai
```

Filter by capabilities:
```bash
go run main.go --provider openai --reasoning
go run main.go --provider openai --vision
```

Sort by cost:
```bash
go run main.go --provider openai --sort cost
```

Export as JSON:
```bash
go run main.go --provider openai --format json > models.json
```

### 3. Get Model Information

```bash
cd examples/client-usage/model-info
go run main.go --model "gpt-4o"
```

Export model configuration:
```bash
go run main.go --model "gpt-4o" --export > model-config.json
```

### 4. Find Models by Criteria

Non-interactive search:
```bash
cd examples/client-usage/find-models
go run main.go --max-cost 1.0 --min-context 100000
go run main.go --reasoning --vision
```

Interactive mode:
```bash
go run main.go --interactive
```

Compare specific models:
```bash
go run main.go --compare "gpt-4o,claude-3-opus"
```

### 5. Calculate API Costs

Single model calculation:
```bash
cd examples/integration/cost-calculator
go run main.go --model "gpt-4o" --input 1000 --output 500
```

Compare multiple models:
```bash
go run main.go --compare "gpt-4o,claude-3-opus" --input 1000 --output 500
```

With caching:
```bash
go run main.go --model "gpt-4o" --input 1000 --output 500 --cached 0.5
```

Batch processing:
```bash
# Create scenarios.json
cat > scenarios.json <<EOF
[
  {
    "model": "gpt-4o",
    "input_tokens": 1000,
    "output_tokens": 500,
    "cached_ratio": 0.5
  },
  {
    "model": "claude-3-opus",
    "input_tokens": 1000,
    "output_tokens": 500,
    "cached_ratio": 0.0
  }
]
EOF

go run main.go --batch scenarios.json --format csv
```

### 6. Interactive Model Selector

```bash
cd examples/integration/model-selector
go run main.go
```

The wizard will guide you through questions about:
- Budget constraints
- Context window requirements
- Reasoning capabilities
- Vision/multimodal support

### 7. Chat Bot (Demo)

```bash
cd examples/integration/chat-bot
go run main.go --provider openai --model gpt-4o
```

**Note**: The chat-bot example demonstrates the UI and integration patterns. For a fully functional chat bot, implement the API call logic shown in the comments in `main.go`.

## Detailed Usage

### Client Usage Examples

#### list-providers

Lists all available AI providers with summary information.

**Features:**
- List all providers from catwalk service
- Show provider name, ID, type, and model count
- Filter by provider type
- Output formats: table, JSON

**Key Concepts:**
- Using `catwalk.New()` client
- Calling `GetProviders()` with ETag support
- Handling `ErrNotModified`
- Formatting output

**Usage:**
```bash
go run main.go                    # List all providers
go run main.go --type openai       # List OpenAI providers only
go run main.go --format json       # Output in JSON
go run main.go --help             # Show help
```

#### list-models

Lists models from a specific provider.

**Features:**
- List all models from a specified provider
- Filter by capabilities (reasoning, vision)
- Sort by cost, context window, or name
- Output formats: table, JSON, CSV

**Key Concepts:**
- Filtering providers by ID
- Iterating through models
- Using model metadata
- Output formatting

**Usage:**
```bash
go run main.go --provider openai                    # List all OpenAI models
go run main.go --provider anthropic --reasoning       # List reasoning models only
go run main.go --provider openai --sort cost          # Sort by cost
go run main.go --provider openai --format json        # Output in JSON
go run main.go --provider openai --format csv         # Output in CSV
```

#### model-info

Displays detailed information about a specific model.

**Features:**
- Display complete model configuration
- Show pricing breakdown (cached/uncached)
- Display reasoning levels and default settings
- Export config as JSON

**Key Concepts:**
- Finding specific models
- Displaying complex nested data
- Exporting data

**Usage:**
```bash
go run main.go --model "gpt-4o"                     # Show model info
go run main.go --model "claude-3-opus" --provider anthropic  # Specify provider
go run main.go --model "gpt-4o" --export              # Export as JSON
```

#### find-models

Finds models matching specific criteria.

**Features:**
- Search models across all providers
- Filter by: max cost, min context window, reasoning support, vision support
- Interactive mode for step-by-step filtering
- Compare multiple models side-by-side
- Ranked list with match scores

**Key Concepts:**
- Multi-provider filtering
- Complex query logic
- Interactive CLI (using bubbletea)
- Scoring and ranking algorithms

**Usage:**
```bash
go run main.go --max-cost 1.0 --min-context 100000       # Non-interactive
go run main.go --reasoning --vision                         # Filter by capabilities
go run main.go --interactive                                # Interactive mode
go run main.go --compare "gpt-4o,claude-3-opus"          # Compare models
```

### Integration Examples

#### cost-calculator

Estimates AI API costs for different models.

**Features:**
- Calculate costs for input/output token estimates
- Compare costs across multiple models
- Account for prompt caching discounts
- Batch calculations (multiple scenarios)
- Export cost comparison as CSV/JSON

**Key Concepts:**
- Using model pricing data
- Cost calculation logic
- Caching discount calculations
- Batch processing

**Usage:**
```bash
go run main.go --model "gpt-4o" --input 1000 --output 500
go run main.go --compare "gpt-4o,claude-3-opus" --input 1000 --output 500
go run main.go --model "gpt-4o" --input 1000 --output 500 --cached 0.5
go run main.go --batch scenarios.json --format csv
```

#### model-selector

Interactive wizard to select the best model based on requirements.

**Features:**
- Wizard-style interface for requirements gathering
- Questions about: budget, context size, reasoning needs, vision support
- Recommends top 3 models with trade-offs explained
- Side-by-side comparison

**Key Concepts:**
- Interactive wizard pattern
- Multi-criteria decision making
- Trade-off analysis

**Usage:**
```bash
go run main.go    # Start interactive wizard
```

#### chat-bot

Interactive CLI chat bot that uses catwalk to select models.

**Features:**
- Interactive chat interface with bubbletea
- Auto-select provider/model based on user requirements
- Display cost estimates before sending messages
- Support for reasoning levels (where applicable)
- Session history with export capability

**Key Concepts:**
- Integrating catwalk with AI API calls
- Real-time cost estimation
- Model selection based on use case
- Interactive CLI with history

**Usage:**
```bash
go run main.go --provider openai --model gpt-4o           # Start with specific model
go run main.go --auto-select                               # Auto-select model
go run main.go --reasoning high --provider anthropic           # With reasoning level
```

**Note**: This example demonstrates UI patterns. For a fully functional chat bot, implement the API call logic shown in the comments.

## Building Examples

All examples can be built and run directly:

```bash
cd examples/client-usage/list-providers
go build -o list-providers
./list-providers
```

Or run directly:
```bash
go run main.go [options]
```

## Environment Variables

All examples respect these environment variables:

- `CATWALK_URL` - URL of the catwalk service (default: http://localhost:8080)

Provider-specific API keys (for integration examples):
- `OPENAI_API_KEY` - For OpenAI provider
- `ANTHROPIC_API_KEY` - For Anthropic provider
- `GEMINI_API_KEY` - For Google Gemini provider
- `XAI_API_KEY` - For xAI/Grok provider
- `DEEPSEEK_API_KEY` - For DeepSeek provider
- `GROQ_API_KEY` - For Groq provider
- And other provider-specific keys as needed

## Dependencies

All examples use the Charm ecosystem for polished CLI experiences:

- `github.com/charmbracelet/bubbletea` - Interactive CLI framework
- `github.com/charmbracelet/lipgloss` - Styling and formatting
- `github.com/charmbracelet/bubbles` - Reusable UI components

For API calls in integration examples:
- `github.com/sashabaranov/go-openai` - OpenAI-compatible API client

These are already included in the project's `go.mod`.

## Contributing

When adding new examples:

1. Follow the existing directory structure
2. Include comprehensive comments explaining:
   - Purpose of the example
   - Key concepts demonstrated
   - Usage instructions
3. Support common flags: `--help`
4. Respect `CATWALK_URL` environment variable
5. Use Charm ecosystem tools for consistent UX
6. Add documentation to this README

## Troubleshooting

### "Connection refused" error

Make sure the catwalk server is running:
```bash
# From project root
go run main.go
```

### "Provider not found" error

Use `list-providers` to see available providers:
```bash
cd examples/client-usage/list-providers
go run main.go
```

### "Model not found" error

Use `list-models` to see available models:
```bash
cd examples/client-usage/list-models
go run main.go --provider <provider-id>
```

### API key errors

Set the appropriate environment variable for your provider:
```bash
export OPENAI_API_KEY=your-key-here
# Windows PowerShell
$env:OPENAI_API_KEY = "your-key-here"
```

## License

MIT License - Part of [Charm](https://charm.land)
