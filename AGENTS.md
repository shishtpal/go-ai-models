# Catwalk - AI Agents/Models Documentation

This document provides comprehensive information about all AI inference providers and models supported by Catwalk, the AI provider database for [Crush](https://github.com/charmbracelet/crush).

## Overview

Catwalk serves as a centralized database of AI model configurations, providing standardized metadata for 21+ inference providers and 200+ models. Each provider configuration includes:

- Model identifiers and names
- Pricing information (input/output, cached/uncached)
- Context window sizes
- Maximum token limits
- Reasoning capabilities
- Vision/multimodal support
- Provider-specific options

---

## Supported Provider Types

Catwalk supports the following API types:

| Type | Description |
|------|-------------|
| `openai` | OpenAI-compatible API (standard) |
| `openai-compat` | OpenAI-compatible API (third-party) |
| `anthropic` | Anthropic Claude API |
| `google` | Google Gemini API |
| `azure` | Azure OpenAI Service |
| `bedrock` | AWS Bedrock API |
| `google-vertex` | Google Vertex AI API |
| `openrouter` | OpenRouter aggregation API |
| `vercel` | Vercel AI SDK API |

---

## Provider Catalog

### 1. OpenAI

**API Type:** `openai`  
**Environment Variables:** `OPENAI_API_KEY`, `OPENAI_API_ENDPOINT`  
**Default Large Model:** `gpt-5.1-codex`  
**Default Small Model:** `gpt-4o`

#### Models

| Model ID | Name | Input Cost | Output Cost | Context | Max Tokens | Reasoning | Vision |
|----------|------|-----------|------------|---------|-----------|-----------|--------|
| `gpt-5.2` | GPT-5.2 | $1.75 | $14.00 | 400K | 128K | Yes | Yes |
| `gpt-5.2-codex` | GPT-5.2 Codex | $1.75 | $14.00 | 400K | 128K | Yes | Yes |
| `gpt-5.1` | GPT-5.1 | $1.25 | $10.00 | 400K | 128K | Yes | Yes |
| `gpt-5.1-codex` | GPT-5.1 Codex | $1.25 | $10.00 | 400K | 128K | Yes | Yes |
| `gpt-5.1-codex-max` | GPT-5.1 Codex Max | $1.25 | $10.00 | 400K | 128K | Yes | Yes |
| `gpt-5.1-codex-mini` | GPT-5.1 Codex Mini | $0.25 | $2.00 | 400K | 128K | Yes | Yes |
| `gpt-5-codex` | GPT-5 Codex | $1.25 | $10.00 | 400K | 128K | Yes | Yes |
| `gpt-5` | GPT-5 | $1.25 | $10.00 | 400K | 128K | Yes | Yes |
| `gpt-5-mini` | GPT-5 Mini | $0.25 | $2.00 | 400K | 128K | Yes | Yes |
| `gpt-5-nano` | GPT-5 Nano | $0.05 | $0.40 | 400K | 128K | Yes | Yes |
| `o4-mini` | o4 Mini | $1.10 | $4.40 | 200K | 50K | Yes | Yes |
| `o3` | o3 | $2.00 | $8.00 | 200K | 50K | Yes | Yes |
| `gpt-4.1` | GPT-4.1 | $2.00 | $8.00 | 1M | 16K | No | Yes |
| `gpt-4.1-mini` | GPT-4.1 Mini | $0.40 | $1.60 | 1M | 16K | No | Yes |
| `gpt-4.1-nano` | GPT-4.1 Nano | $0.10 | $0.40 | 1M | 16K | No | Yes |
| `o3-mini` | o3 Mini | $1.10 | $4.40 | 200K | 50K | Yes | No |
| `gpt-4o` | GPT-4o | $2.50 | $10.00 | 128K | 8K | No | Yes |
| `gpt-4o-mini` | GPT-4o-mini | $0.15 | $0.60 | 128K | 8K | No | Yes |

**Features:**
- Reasoning levels: minimal, low, medium, high
- Prompt caching supported (90% discount on cached tokens)
- Vision/multimodal support on most models
- Codex variants optimized for code generation

---

### 2. Anthropic

**API Type:** `anthropic`  
**Environment Variables:** `ANTHROPIC_API_KEY`, `ANTHROPIC_API_ENDPOINT`  
**Default Large Model:** `claude-sonnet-4-5-20250929`  
**Default Small Model:** `claude-3-5-haiku-20241022`

#### Models

| Model ID | Name | Input Cost | Output Cost | Context | Max Tokens | Reasoning | Vision |
|----------|------|-----------|------------|---------|-----------|-----------|--------|
| `claude-sonnet-4-5-20250929` | Claude Sonnet 4.5 | $3.00 | $15.00 | 200K | 50K | Yes | Yes |
| `claude-opus-4-6` | Claude Opus 4.6 | $5.00 | $25.00 | 200K | 126K | Yes | Yes |
| `claude-opus-4-5-20251101` | Claude Opus 4.5 | $5.00 | $25.00 | 200K | 50K | Yes | Yes |
| `claude-haiku-4-5-20251001` | Claude 4.5 Haiku | $1.00 | $5.00 | 200K | 32K | Yes | Yes |
| `claude-opus-4-1-20250805` | Claude Opus 4.1 | $15.00 | $75.00 | 200K | 32K | Yes | Yes |
| `claude-opus-4-20250514` | Claude Opus 4 | $15.00 | $75.00 | 200K | 32K | Yes | Yes |
| `claude-sonnet-4-20250514` | Claude Sonnet 4 | $3.00 | $15.00 | 200K | 50K | Yes | Yes |
| `claude-3-7-sonnet-20250219` | Claude 3.7 Sonnet | $3.00 | $15.00 | 200K | 50K | Yes | Yes |
| `claude-3-5-haiku-20241022` | Claude 3.5 Haiku | $0.80 | $4.00 | 200K | 5K | No | Yes |
| `claude-3-5-sonnet-20240620` | Claude 3.5 Sonnet (Old) | $3.00 | $15.00 | 200K | 5K | No | Yes |
| `claude-3-5-sonnet-20241022` | Claude 3.5 Sonnet (New) | $3.00 | $15.00 | 200K | 5K | No | Yes |

**Features:**
- Extended thinking/reasoning capabilities
- Strong coding and analysis capabilities
- Vision support on all models
- Prompt caching supported

---

### 3. Google Gemini

**API Type:** `google`  
**Environment Variables:** `GEMINI_API_KEY`, `GEMINI_API_ENDPOINT`  
**Default Large Model:** `gemini-2.5-pro`  
**Default Small Model:** `gemini-2.5-flash`

#### Models

| Model ID | Name | Input Cost | Output Cost | Context | Max Tokens | Reasoning | Vision |
|----------|------|-----------|------------|---------|-----------|-----------|--------|
| `gemini-3-pro-preview` | Gemini 3 Pro (Preview) | $2.00 | $12.00 | 1M | 64K | Yes | Yes |
| `gemini-3-flash-preview` | Gemini 3 Flash (Preview) | $0.50 | $3.00 | 1M | 50K | Yes | Yes |
| `gemini-2.5-pro` | Gemini 2.5 Pro | $1.25 | $10.00 | 1M | 50K | Yes | Yes |
| `gemini-2.5-flash` | Gemini 2.5 Flash | $0.30 | $2.50 | 1M | 50K | Yes | Yes |

**Features:**
- 1M token context window (largest in the industry)
- Strong multimodal capabilities
- Fast inference with Flash models
- Reasoning capabilities on all models

---

### 4. xAI (Grok)

**API Type:** `openai-compat`  
**Environment Variables:** `XAI_API_KEY`  
**API Endpoint:** `https://api.x.ai/v1`  
**Default Large Model:** `grok-code-fast`  
**Default Small Model:** `grok-3-mini`

#### Models

| Model ID | Name | Input Cost | Output Cost | Context | Max Tokens | Reasoning | Vision |
|----------|------|-----------|------------|---------|-----------|-----------|--------|
| `grok-code-fast` | Grok Code Fast | $0.20 | $1.50 | 256K | 20K | Yes | No |
| `grok-4-1-fast-reasoning` | Grok 4.1 Fast Reasoning | $0.20 | $0.50 | 2M | 200K | Yes | Yes |
| `grok-4-1-fast-non-reasoning` | Grok 4.1 Fast Non-Reasoning | $0.20 | $0.50 | 2M | 200K | No | Yes |
| `grok-4` | Grok 4 | $3.00 | $15.00 | 256K | 20K | Yes | Yes |
| `grok-4-fast` | Grok 4 Fast | $0.20 | $0.50 | 2M | 200K | Yes | Yes |
| `grok-3-mini` | Grok 3 Mini | $0.30 | $0.50 | 131K | 20K | Yes | No |
| `grok-3` | Grok 3 | $3.00 | $15.00 | 131K | 20K | No | No |

**Features:**
- Up to 2M token context window
- Code-optimized models
- Reasoning levels: low, high
- Fast variants for reduced latency

---

### 5. DeepSeek

**API Type:** `openai-compat`  
**Environment Variables:** `DEEPSEEK_API_KEY`  
**API Endpoint:** `https://api.deepseek.com/v1`  
**Default Large Model:** `deepseek-reasoner`  
**Default Small Model:** `deepseek-chat`

#### Models

| Model ID | Name | Input Cost | Output Cost | Context | Max Tokens | Reasoning | Vision |
|----------|------|-----------|------------|---------|-----------|-----------|--------|
| `deepseek-chat` | DeepSeek-V3.2 (Non-thinking) | $0.28 | $0.42 | 128K | 4K | No | No |
| `deepseek-reasoner` | DeepSeek-V3.2 (Thinking) | $0.28 | $0.42 | 128K | 32K | Yes | No |

**Features:**
- Extremely cost-effective
- Reasoning levels: low, medium, high
- Thinking mode for complex reasoning tasks

---

### 6. Groq

**API Type:** `openai-compat`  
**Environment Variables:** `GROQ_API_KEY`  
**API Endpoint:** `https://api.groq.com/openai/v1`  
**Default Large Model:** `moonshotai/kimi-k2-instruct-0905`  
**Default Small Model:** `qwen/qwen3-32b`

#### Models

| Model ID | Name | Input Cost | Output Cost | Context | Max Tokens | Reasoning | Vision |
|----------|------|-----------|------------|---------|-----------|-----------|--------|
| `moonshotai/kimi-k2-instruct-0905` | Kimi K2 0905 | $1.00 | $3.00 | 131K | 10K | Yes | No |
| `qwen/qwen3-32b` | Qwen3 32B | $0.29 | $0.59 | 131K | 10K | No | No |

**Features:**
- Ultra-fast inference (Groq's LPU architecture)
- Low latency, high throughput
- Reasoning levels: low, medium, high

---

### 7. OpenRouter

**API Type:** `openrouter`  
**Environment Variables:** `OPENROUTER_API_KEY`  
**API Endpoint:** `https://openrouter.ai/api/v1`  
**Default Large Model:** `anthropic/claude-sonnet-4`  
**Default Small Model:** `anthropic/claude-3.5-haiku`

OpenRouter is an aggregation service providing access to 100+ models from multiple providers. Models include:

- Anthropic Claude models
- OpenAI GPT models
- Google Gemini models
- Meta Llama models
- Mistral models
- And many more

**Features:**
- Unified API for multiple providers
- Automatic endpoint selection based on uptime and capabilities
- Transparent pricing comparison
- Tool/function calling support filtering
- Context window and capability filtering

---

### 8. AWS Bedrock

**API Type:** `bedrock`  
**Environment Variables:** Uses AWS credentials  
**Default Large Model:** `anthropic.claude-sonnet-4-5-20250929-v1:0`  
**Default Small Model:** `anthropic.claude-3-5-haiku-20241022-v1:0`

#### Models

| Model ID | Name | Input Cost | Output Cost | Context | Max Tokens | Reasoning | Vision |
|----------|------|-----------|------------|---------|-----------|-----------|--------|
| `anthropic.claude-sonnet-4-5-20250929-v1:0` | AWS Claude Sonnet 4.5 | $3.00 | $15.00 | 200K | 50K | Yes | Yes |
| `anthropic.claude-haiku-4-5-20251001-v1:0` | AWS Claude Haiku 4.5 | $0.80 | $4.00 | 200K | 50K | No | Yes |
| `anthropic.claude-opus-4-5-20251101-v1:0` | AWS Claude Opus 4.5 | $5.00 | $25.00 | 200K | 50K | Yes | Yes |
| `anthropic.claude-opus-4-1-20250805-v1:0` | AWS Claude Opus 4.1 | $15.00 | $75.00 | 200K | 50K | Yes | Yes |
| `anthropic.claude-opus-4-20250514-v1:0` | AWS Claude Opus 4 | $15.00 | $75.00 | 200K | 50K | Yes | Yes |
| `anthropic.claude-sonnet-4-20250514-v1:0` | AWS Claude Sonnet 4 | $3.00 | $15.00 | 200K | 50K | Yes | Yes |
| `anthropic.claude-3-7-sonnet-20250219-v1:0` | AWS Claude 3.7 Sonnet | $3.00 | $15.00 | 200K | 50K | Yes | Yes |
| `anthropic.claude-3-5-haiku-20241022-v1:0` | AWS Claude 3.5 Haiku | $0.80 | $4.00 | 200K | 50K | No | Yes |

**Features:**
- AWS infrastructure integration
- Same models as Anthropic but hosted on AWS
- Enterprise compliance and security features
- Regional deployment options

---

### 9. Azure OpenAI

**API Type:** `azure`  
**Environment Variables:** Azure-specific credentials  
**Default Large Model:** `gpt-5.1-codex`  
**Default Small Model:** `gpt-4o`

Azure OpenAI provides the same models as OpenAI but hosted on Microsoft Azure infrastructure with enterprise features.

---

### 10. Google Vertex AI

**API Type:** `google-vertex`  
**Environment Variables:** Google Cloud credentials  
**Default Large Model:** `gemini-2.5-pro`  
**Default Small Model:** `gemini-2.5-flash`

Vertex AI provides Google's Gemini models with enterprise features, regional deployment, and integration with Google Cloud services.

---

### 11. Cerebras

**API Type:** `openai-compat`  
**Environment Variables:** `CEREBRAS_API_KEY`  
**API Endpoint:** `https://api.cerebras.ai/v1`  
**Default Large Model:** `gpt-oss-120b`  
**Default Small Model:** `qwen-3-32b`

#### Models

| Model ID | Name | Input Cost | Output Cost | Context | Max Tokens | Reasoning | Vision |
|----------|------|-----------|------------|---------|-----------|-----------|--------|
| `llama-3.3-70b` | Llama 3.3 70B | $0.85 | $1.20 | 131K | 25K | No | No |
| `gpt-oss-120b` | OpenAI GPT OSS | $0.35 | $0.75 | 131K | 25K | Yes | No |
| `qwen-3-32b` | Qwen 3 32B | $0.40 | $0.80 | 131K | 25K | No | No |
| `qwen-3-235b-a22b-instruct-2507` | Qwen 3 235B Instruct | $0.60 | $1.20 | 131K | 25K | No | No |
| `zai-glm-4.7` | Z.ai GLM 4.7 | $2.25 | $2.75 | 131K | 25K | No | No |

**Features:**
- Ultra-fast inference on Cerebras CS-2 systems
- Large model support (up to 235B parameters)
- Reasoning levels: low, medium, high

---

### 12. Hugging Face

**API Type:** `openai-compat`  
**Environment Variables:** `HF_TOKEN`  
**API Endpoint:** `https://router.huggingface.co/v1`  
**Default Large Model:** `moonshotai/Kimi-K2-Instruct-0905:groq`  
**Default Small Model:** `openai/gpt-oss-20b:groq`

Hugging Face provides access to open-source models hosted on their inference platform, with routing to multiple backend providers (Groq, Cerebras, Fireworks AI).

#### Models Include:
- Qwen models (3, 30B, 235B, Coder variants)
- DeepSeek models
- Meta Llama models (3.1, 3.3, 4)
- OpenAI GPT OSS models
- MoonshotAI Kimi models
- Z.ai GLM models

**Features:**
- Access to 100+ open-source models
- Multiple backend providers (Groq, Cerebras, Fireworks AI)
- Cost-effective open-source alternatives
- Specialized coding models

---

### 13. GitHub Copilot

**API Type:** `openai-compat`  
**API Endpoint:** `https://api.githubcopilot.com`  
**Default Large Model:** `claude-sonnet-4.5`  
**Default Small Model:** `claude-haiku-4.5`

GitHub Copilot provides access to multiple models at no additional cost for Copilot subscribers:

- Claude models (Haiku 4.5, Opus 4.5, Sonnet 4, 4.5)
- Gemini models (2.5 Pro, 3 Flash/Pro Preview)
- GPT models (3.5 Turbo, 4, 4 Turbo, 5, 5.1, 5.2, and Codex variants)
- Grok Code Fast 1

**Features:**
- Free with GitHub Copilot subscription
- Multiple provider access
- No per-token billing (subscription-based)
- Limited max tokens (16K-64K)

---

### 14. Vercel AI SDK

**API Type:** `vercel`  
**Environment Variables:** Vercel-specific credentials  
**Default Large Model:** `claude-sonnet-4-5-20250929`  
**Default Small Model:** `claude-3-5-haiku-20241022`

Vercel provides access to multiple AI models through their AI SDK, optimized for web applications.

---

### 15. Additional Providers

#### Kimi Coding
- Provider: MoonshotAI
- Models: Kimi K2 series
- Specialization: Coding tasks

#### MiniMax
- Provider: MiniMax
- API Type: `openai-compat`
- Environment: `MINIMAX_API_KEY`

#### Venice
- Provider: Venice
- API Type: `openai-compat`
- Environment: `VENICE_API_KEY`

#### Chutes
- Provider: Chutes
- API Type: `openai-compat`
- Environment: `CHUTES_API_KEY`

#### AIHubMix
- Provider: AIHubMix
- API Type: `openai-compat`
- Environment: `AIHUBMIX_API_KEY`

#### Zai
- Provider: Z.ai
- API Type: `openai-compat`
- Environment: `ZAI_API_KEY`

#### Synthetic
- Test provider for development and testing

---

## Model Capabilities Reference

### Reasoning Levels

Some models support configurable reasoning effort levels:

| Level | Description |
|-------|-------------|
| `minimal` | Fastest inference, minimal reasoning |
| `low` | Quick reasoning for simple tasks |
| `medium` | Balanced reasoning (default) |
| `high` | Maximum reasoning for complex tasks |

### Prompt Caching

Providers that support prompt caching offer significant cost savings:

- **OpenAI**: 90% discount on cached tokens
- **Anthropic**: Cache read/write pricing
- **Google**: Cache read pricing

### Vision/Multimodal Support

Models with vision capabilities can process images and other media:

- **Full Vision**: OpenAI GPT series, Anthropic Claude, Google Gemini, xAI Grok 4+
- **Text Only**: Most smaller models and specialized coding models

### Context Window Sizes

| Size Range | Models |
|------------|--------|
| 1M+ tokens | Google Gemini (1M), OpenAI GPT-4.1 (1M) |
| 400K tokens | OpenAI GPT-5 series |
| 256K-200K tokens | Anthropic Claude, xAI Grok, DeepSeek |
| 131K tokens | Groq, Cerebras, Hugging Face models |
| 128K tokens | OpenAI GPT-4o, older models |

---

## Cost Comparison

### Most Cost-Effective (per 1M input tokens)

1. **GPT-5 Nano** - $0.05
2. **GPT-5.1 Codex Mini** - $0.25
3. **GPT-5 Mini** - $0.25
4. **DeepSeek** - $0.28
5. **GPT-4.1 Nano** - $0.10

### Premium/High-Performance

1. **Claude Opus 4/4.1** - $15.00
2. **GPT-5.2** - $1.75
3. **Claude Opus 4.5/4.6** - $5.00
4. **GPT-5.1** - $1.25

---

## Usage Example

```go
import "charm.land/catwalk/pkg/catwalk"

client := catwalk.New()
providers, err := client.GetProviders(ctx, etag)
if err != nil {
    // handle error
}

for _, provider := range providers {
    fmt.Printf("Provider: %s (%d models)\n", provider.Name, len(provider.Models))
    for _, model := range provider.Models {
        fmt.Printf("  - %s: $%.2f in, $%.2f out\n",
            model.Name, model.CostPer1MIn, model.CostPer1MOut)
    }
}
```

---

## Contributing

To add a new provider:

1. Create `internal/providers/configs/{provider}.json`
2. Add provider function in `internal/providers/providers.go`
3. Create CLI generator in `cmd/{provider}/main.go` if API available
4. Update this documentation

See `CRUSH.md` for detailed guidelines.

---

## License

MIT License - Part of [Charm](https://charm.land)
