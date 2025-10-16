# Turing Roulette

A game where you create riddles to challenge AI models. Win by stumping some models but not all. Features difficulty ratings, statistics tracking, and a global leaderboard.

## Features

- Real-time streaming responses from multiple AI providers
- Difficulty-based scoring system (Easy/Medium/Hard)
- Persistent statistics tracking
- Global leaderboard (top 100 scores)
- Support for OpenAI, Anthropic, Google, Ollama, and HuggingFace models
- Configuration-driven model selection

## Project Structure

```
turing-roulette/
├── cmd/server/main.go     # Go backend server
├── go.mod                 # Go module dependencies
├── config.template.json   # Configuration template (safe to commit)
├── config.json            # Model configuration (gitignored, create from template)
├── stats.json             # Statistics data (auto-generated)
├── leaderboard.json       # Leaderboard data (auto-generated)
├── frontend/              # React frontend source
├── static/                # Built frontend assets
│   └── index.html         # React application
├── .gitignore             # Git ignore rules
└── README.md
```

## Prerequisites

- Go 1.19 or higher
- Node.js 16 or higher (for React frontend development)
- Optional: Ollama for local model testing
- Optional: API keys for OpenAI, Anthropic, Google, or HuggingFace

## Installation

### 1. Backend Setup

Clone the repository and install Go dependencies:

```bash
git clone <repository-url>
cd turing-roulette
go mod init turing-roulette
go get github.com/gorilla/websocket
```

### 2. API Keys Setup (IMPORTANT)

**Never commit API keys to version control!** Use environment variables instead.

#### Quick Setup

1. Copy the environment template:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and add your actual API keys.

3. Set the environment variables in your shell (see examples below).

#### Setting Environment Variables

For Linux/macOS:
```bash
export OPENAI_API_KEY="sk-your-openai-key"
export ANTHROPIC_API_KEY="sk-ant-your-anthropic-key"
export GOOGLE_API_KEY="your-google-key"
export HUGGINGFACE_API_KEY="hf_your-huggingface-token"
```

For Windows:
```cmd
set OPENAI_API_KEY=sk-your-openai-key
set ANTHROPIC_API_KEY=sk-ant-your-anthropic-key
set GOOGLE_API_KEY=your-google-key
set HUGGINGFACE_API_KEY=hf_your-huggingface-token
```

For PowerShell:
```powershell
$env:OPENAI_API_KEY="sk-your-openai-key"
$env:ANTHROPIC_API_KEY="sk-ant-your-anthropic-key"
$env:GOOGLE_API_KEY="your-google-key"
$env:HUGGINGFACE_API_KEY="hf_your-huggingface-token"
```

### 3. Configuration

Copy `config.template.json` to `config.json` and customize as needed. The template includes placeholder values that will be overridden by environment variables.

Choose one of the following configurations:

#### Option A: Ollama (Free, Local Testing)

```json
{
  "models": [
    {
      "name": "Llama 2",
      "provider": "ollama",
      "model": "llama2",
      "endpoint": "http://localhost:11434"
    },
    {
      "name": "Mistral",
      "provider": "ollama",
      "model": "mistral",
      "endpoint": "http://localhost:11434"
    },
    {
      "name": "CodeLlama",
      "provider": "ollama",
      "model": "codellama",
      "endpoint": "http://localhost:11434"
    }
  ]
}
```

Install Ollama and pull models:

```bash
# Install Ollama from https://ollama.ai
curl -fsSL https://ollama.ai/install.sh | sh

# Pull models
ollama pull llama2
ollama pull mistral
ollama pull codellama
```

#### Option B: Production APIs

```json
{
  "models": [
    {
      "name": "GPT-4",
      "provider": "openai",
      "model": "gpt-4-turbo-preview",
      "apiKey": "sk-your-openai-api-key"
    },
    {
      "name": "Claude",
      "provider": "anthropic",
      "model": "claude-3-sonnet-20240229",
      "apiKey": "sk-ant-your-anthropic-api-key"
    },
    {
      "name": "Gemini",
      "provider": "google",
      "model": "gemini-pro",
      "apiKey": "your-google-api-key"
    },
    {
      "name": "Llama on HF",
      "provider": "huggingface",
      "model": "meta-llama/Llama-2-7b-chat-hf",
      "apiKey": "hf_your-huggingface-token"
    }
  ]
}
```

#### Option C: Mixed Configuration

You can mix local and cloud models:

```json
{
  "models": [
    {
      "name": "GPT-4",
      "provider": "openai",
      "model": "gpt-4-turbo-preview",
      "apiKey": "sk-..."
    },
    {
      "name": "Local Llama",
      "provider": "ollama",
      "model": "llama2",
      "endpoint": "http://localhost:11434"
    }
  ]
}
```

### 4. Run the Server

```bash
# Copy the configuration template
cp config.template.json config.json

# Run the server (ensure environment variables are set)
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### 5. Frontend Setup

For development with React:

```bash
# Create React app
npx create-react-app frontend
cd frontend

# Install dependencies
npm install lucide-react

# Install Tailwind CSS
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

Configure `tailwind.config.js`:

```javascript
module.exports = {
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  theme: { extend: {} },
  plugins: [],
}
```

Update `src/index.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

Replace `src/App.js` with the React component and run:

```bash
npm start
```

For production, build and serve from the Go server:

```bash
cd frontend
npm run build
cp -r build/* ../static/
```

## Configuration Reference

### Model Configuration Fields

- `name`: Display name for the model (shown in UI)
- `provider`: AI provider type
  - `openai`: OpenAI GPT models
  - `anthropic`: Anthropic Claude models
  - `google`: Google Gemini models
  - `ollama`: Local Ollama models
  - `huggingface`: HuggingFace Inference API
- `model`: Model identifier specific to provider
- `apiKey`: API authentication key (not needed for Ollama)
- `endpoint`: Custom endpoint URL (optional, mainly for Ollama)

### Provider-Specific Configuration

#### OpenAI

- Models: `gpt-4-turbo-preview`, `gpt-4`, `gpt-3.5-turbo`
- API Key: Get from https://platform.openai.com/api-keys
- Documentation: https://platform.openai.com/docs

#### Anthropic (Claude)

- Models: `claude-3-opus-20240229`, `claude-3-sonnet-20240229`, `claude-3-haiku-20240307`
- API Key: Get from https://console.anthropic.com/
- Documentation: https://docs.anthropic.com/

#### Google (Gemini)

- Models: `gemini-pro`, `gemini-pro-vision`
- API Key: Get from https://makersuite.google.com/app/apikey
- Documentation: https://ai.google.dev/docs

#### Ollama (Local)

- Models: Any model from https://ollama.ai/library
- Popular choices: `llama2`, `mistral`, `codellama`, `vicuna`, `orca-mini`
- No API key required
- Default endpoint: `http://localhost:11434`
- Documentation: https://github.com/ollama/ollama

#### HuggingFace

- Models: Any text-generation model on HuggingFace Hub
- Popular choices: `meta-llama/Llama-2-7b-chat-hf`, `mistralai/Mistral-7B-Instruct-v0.1`
- API Token: Get from https://huggingface.co/settings/tokens
- Documentation: https://huggingface.co/docs/api-inference/

## Game Rules

### Objective

Win by creating a riddle that stumps some AI models but not all of them.

### Gameplay

1. Select a difficulty level (Easy, Medium, Hard)
2. Create a riddle with an answer
3. Provide 3 clues that will be revealed progressively
4. AI models attempt to solve the riddle
5. Incorrect models receive one clue per round
6. Game ends when all models are correct or all clues are exhausted

### Win Conditions

- You WIN if: Some models guess correctly, but not all
- You LOSE if: All models guess correctly, or none guess correctly

### Scoring System

Base score: 100 points

Multipliers:
- Easy: 1.0x
- Medium: 1.5x
- Hard: 2.0x

Bonuses:
- Speed bonus: Up to 50 points (for completion under 60 seconds)
- Stump bonus: 20 points per model stumped

Formula:
```
Score = (100 * difficulty_multiplier) + speed_bonus + (stumped_models * 20)
```

## Statistics Tracking

The game automatically tracks:

- Total games played
- Wins and losses
- Win rate percentage
- Games by difficulty level
- Average game duration
- Total playtime

Statistics are saved to `stats.json` and persist between sessions.

## Leaderboard

- Top 100 highest scoring games
- Displays: Rank, Score, Difficulty, Result, Riddle preview, Duration
- Automatically updated after each game
- Saved to `leaderboard.json`

## API Endpoints

### WebSocket

- `ws://localhost:8080/ws` - Game communication channel

### HTTP

- `GET /config` - Returns current model configuration
- `GET /stats` - Returns player statistics
- `GET /leaderboard` - Returns top 100 scores

## Cost Estimates

### Free Options

- Ollama: Completely free, runs locally
  - Requires: 8GB+ RAM per model
  - GPU recommended for faster responses

### Paid APIs (approximate per game)

- OpenAI GPT-4: $0.01 - $0.03
- OpenAI GPT-3.5: $0.001 - $0.003
- Anthropic Claude: $0.01 - $0.02
- Google Gemini: Free tier available, then ~$0.001
- HuggingFace: Free tier available, paid inference ~$0.002

## Troubleshooting

### WebSocket Connection Failed

- Verify server is running on port 8080
- Check firewall settings
- Ensure frontend is configured to connect to correct URL
- For production, update WebSocket URL in frontend code

### Ollama Not Responding

```bash
# Check if Ollama is running
ollama list

# Start Ollama service
ollama serve

# Verify models are installed
ollama list

# Pull missing models
ollama pull llama2
```

### API Authentication Errors

- Verify API keys are correct
- Check API key has sufficient permissions
- Ensure you have available credits/quota
- Check for typos in config.json

### Models Timing Out

- Increase timeout in cmd/server/main.go (currently 60 seconds)
- Use faster models (e.g., GPT-3.5 instead of GPT-4)
- Check network connection
- For HuggingFace, models may need warm-up time

### Statistics Not Saving

- Ensure server has write permissions in directory
- Check disk space
- Verify JSON files are not corrupted
- Delete and regenerate stats.json if necessary

## Development

### Adding New AI Providers

1. Define request/response structures in cmd/server/main.go
2. Implement `stream[Provider]` function following existing patterns
3. Add provider case in `streamModelResponse` function
4. Add provider icon mapping in frontend `getModelIcon` function
5. Update configuration documentation

### Modifying Scoring Algorithm

Edit the `calculateScore` function in cmd/server/main.go:

```go
func calculateScore(result GameResult) int {
    // Customize scoring logic here
    baseScore := 100
    // Add your modifications
    return score
}
```

### Customizing UI

The frontend uses Tailwind CSS. Modify colors and styling in the React component:

- Difficulty colors: `getDifficultyColor` function
- Model colors: `getModelColor` function
- Gradient backgrounds: Update className attributes

## Security Considerations

- Never commit API keys to version control
- Use environment variables for sensitive data in production
- Implement rate limiting for public deployments
- Add authentication for multi-user scenarios
- Sanitize user input for riddles and clues
- Use HTTPS in production environments

## Performance Optimization

### Backend

- Implement response caching for repeated riddles
- Add connection pooling for database operations
- Use goroutine pools to limit concurrent model requests
- Implement request queuing for high traffic

### Frontend

- Lazy load leaderboard data
- Implement pagination for long leaderboards
- Cache statistics locally
- Use React.memo for expensive components

## Deployment

### Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
COPY static ./static
COPY config.json .
EXPOSE 8080
CMD ["./server"]
```

Build and run:

```bash
docker build -t turing-roulette .
docker run -p 8080:8080 -v $(pwd)/config.json:/root/config.json turing-roulette
```

### Cloud Deployment

1. Deploy Go server to cloud platform (AWS, GCP, Azure, Heroku)
2. Configure environment variables for API keys
3. Set up persistent storage for stats.json and leaderboard.json
4. Configure domain and SSL certificate
5. Update WebSocket URL in frontend to production domain

## Contributing

Contributions are welcome. Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License

## Support

For issues and questions:
- Check the Troubleshooting section
- Review provider documentation
- Open an issue on GitHub

## Acknowledgments

- OpenAI for GPT models
- Anthropic for Claude
- Google for Gemini
- Ollama for local model support
- HuggingFace for model hosting
- Gorilla WebSocket library
- React and Tailwind CSS communities


	