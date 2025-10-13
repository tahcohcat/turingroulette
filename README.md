# Turing Roulette

A game where you create riddles to challenge AI models. Win by stumping some models but not all!

## Setup

### Prerequisites

- Go 1.19+ installed
- Node.js 16+ (for React frontend)
- (Optional) Ollama installed for local testing

### Backend Setup

1. **Install Go dependencies:**
```bash
go get github.com/gorilla/websocket
```

2. **Create a configuration file:**

Choose one of these options:

**Option A: Use Ollama (Free, Local Testing)**
```bash
# Install Ollama from https://ollama.ai
# Pull models
ollama pull llama2
ollama pull mistral
ollama pull codellama

# Copy the Ollama config
cp config-ollama.json config.json
```

**Option B: Use Production APIs**
```bash
# Copy the template
cp config.json.example config.json

# Edit config.json and add your API keys
```

3. **Run the server:**
```bash
go run server.go
```

The server will start on `http://localhost:8080`

### Configuration Format

Your `config.json` should look like this:

```json
{
  "models": [
    {
      "name": "Display Name",
      "provider": "openai|anthropic|google|ollama",
      "model": "model-identifier",
      "apiKey": "your-api-key",
      "endpoint": "http://localhost:11434"
    }
  ]
}
```

**Provider-specific settings:**

**OpenAI:**
- `provider`: `"openai"`
- `model`: `"gpt-4-turbo-preview"`, `"gpt-3.5-turbo"`, etc.
- `apiKey`: Your OpenAI API key from https://platform.openai.com/api-keys

**Anthropic (Claude):**
- `provider`: `"anthropic"`
- `model`: `"claude-3-sonnet-20240229"`, `"claude-3-opus-20240229"`, etc.
- `apiKey`: Your Anthropic API key from https://console.anthropic.com/

**Google (Gemini):**
- `provider`: `"google"`
- `model`: `"gemini-pro"`, `"gemini-pro-vision"`, etc.
- `apiKey`: Your Google AI API key from https://makersuite.google.com/app/apikey

**Ollama (Local):**
- `provider`: `"ollama"`
- `model`: `"llama2"`, `"mistral"`, `"codellama"`, etc.
- `endpoint`: `"http://localhost:11434"` (default)
- `apiKey`: Not required

### Frontend Setup

The React component can be integrated into your React app. For standalone testing:

1. **Create a React app (if you don't have one):**
```bash
npx create-react-app turing-roulette-frontend
cd turing-roulette-frontend
```

2. **Install dependencies:**
```bash
npm install lucide-react
```

3. **Setup Tailwind CSS:**
```bash
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

Edit `tailwind.config.js`:
```js
module.exports = {
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  theme: { extend: {} },
  plugins: [],
}
```

4. **Replace `src/App.js` with the React component**

5. **Update `src/index.css`:**
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

6. **Run the frontend:**
```bash
npm start
```

## How to Play

1. **Create a Riddle:** Write an interesting riddle
2. **Set the Answer:** Provide the correct answer
3. **Add Clues:** Create 3 clues that will be revealed progressively
4. **Start Game:** The AIs will attempt to solve your riddle
5. **Win Condition:** You win if SOME but not ALL models guess correctly

## Game Rules

- Round 1: AIs see only the riddle
- Round 2+: Incorrect AIs get one clue per round
- Game ends when:
  - All AIs guess correctly (you lose)
  - All clues are exhausted (check if some were right)
  - You win if some but not all AIs guessed correctly

## Example Configuration Files

### Mixed Configuration (Production + Local)
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
    },
    {
      "name": "Claude",
      "provider": "anthropic",
      "model": "claude-3-sonnet-20240229",
      "apiKey": "sk-ant-..."
    }
  ]
}
```

## Troubleshooting

**WebSocket connection failed:**
- Ensure the Go server is running on port 8080
- Check CORS settings if running frontend on different port

**Ollama not responding:**
- Make sure Ollama service is running: `ollama serve`
- Verify models are installed: `ollama list`
- Check endpoint in config.json matches Ollama's address

**API errors:**
- Verify API keys are correct and have sufficient credits
- Check rate limits on your API accounts
- Review server logs for detailed error messages

## Cost Considerations

- **Ollama**: Free, runs locally (requires ~8GB RAM per model)
- **OpenAI**: ~$0.01-0.03 per riddle with GPT-4
- **Anthropic**: ~$0.015 per riddle with Claude 3
- **Google**: Free tier available, then ~$0.001 per riddle

## Development

To add a new AI provider:

1. Add provider-specific request/response structures
2. Implement a `stream[Provider]` function
3. Add provider case in `streamModelResponse`
4. Update frontend icon mapping

## License

MIT