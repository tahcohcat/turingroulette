# Agent Instructions for Turing Roulette

## Commands
- **Build**: `go build -o turingroulette ./cmd/server` (from repo root)
- **Run**: `go run ./cmd/server/main.go` or `go run cmd/server/main.go`
- **Test (Go)**: `go test ./...` (run all), `go test ./path/to/package` (single package)
- **Frontend build**: `cd frontend && npm run build` (outputs to frontend/build/, copy to static/)
- **Frontend dev**: `cd frontend && npm start` (runs on http://localhost:3000)
- **Frontend test**: `cd frontend && npm test`

## Architecture
- **Backend**: Single Go server (cmd/server/main.go, ~800 lines), WebSocket + HTTP endpoints on :8080
- **Frontend**: React SPA in frontend/ using Tailwind CSS, served from static/ in production
- **Providers**: Integrates OpenAI, Anthropic, Google, Ollama, HuggingFace via streaming APIs
- **Persistence**: JSON files (config.json, stats.json, leaderboard.json) with mutex-protected access
- **Communication**: WebSocket /ws for game, HTTP for /config, /stats, /leaderboard, CORS enabled for localhost:3000

## Code Style (Go)
- **Imports**: Standard lib, then third-party (gorilla/websocket), group separated by blank lines
- **Error handling**: Check errors immediately, log and return early; use log.Fatal for startup errors
- **Naming**: CamelCase exports, camelCase unexported, descriptive names (ModelConfig, GameState)
- **Concurrency**: Use goroutines for parallel model requests, protect shared state with sync.Mutex
- **JSON**: Use json tags on structs, MarshalIndent with 2-space indent for files
- **Types**: Strongly typed structs for all data, use maps for model states/difficulty multipliers
