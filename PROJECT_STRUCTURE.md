# Turing Roulette - Project Structure

## Overview

This document outlines the complete project structure, file organization, and architectural decisions for the Turing Roulette application.

## Directory Structure

```
turing-roulette/
│
├── backend/
│   ├── server.go                 # Main server application
│   ├── go.mod                    # Go module definition
│   ├── go.sum                    # Go dependency checksums
│   ├── config.json               # Model configuration
│   ├── config-ollama.json        # Ollama-specific config example
│   ├── stats.json                # Statistics data (auto-generated)
│   ├── leaderboard.json          # Leaderboard data (auto-generated)
│   │
│   ├── models/                   # AI model integrations (optional refactor)
│   │   ├── openai.go
│   │   ├── anthropic.go
│   │   ├── google.go
│   │   ├── ollama.go
│   │   └── huggingface.go
│   │
│   ├── handlers/                 # HTTP/WebSocket handlers (optional refactor)
│   │   ├── websocket.go
│   │   ├── config.go
│   │   ├── stats.go
│   │   └── leaderboard.go
│   │
│   └── game/                     # Game logic (optional refactor)
│       ├── scoring.go
│       ├── state.go
│       └── validation.go
│
├── frontend/
│   ├── public/
│   │   ├── index.html
│   │   ├── favicon.ico
│   │   └── manifest.json
│   │
│   ├── src/
│   │   ├── App.js                # Main React component
│   │   ├── index.js              # React entry point
│   │   ├── index.css             # Tailwind CSS imports
│   │   │
│   │   ├── components/           # React components (optional refactor)
│   │   │   ├── Menu.js
│   │   │   ├── GameSetup.js
│   │   │   ├── GamePlay.js
│   │   │   ├── GameResults.js
│   │   │   ├── Statistics.js
│   │   │   ├── Leaderboard.js
│   │   │   └── ModelColumn.js
│   │   │
│   │   ├── hooks/                # Custom React hooks (optional)
│   │   │   ├── useWebSocket.js
│   │   │   ├── useGameState.js
│   │   │   └── useStats.js
│   │   │
│   │   └── utils/                # Utility functions (optional)
│   │       ├── api.js
│   │       ├── colors.js
│   │       └── formatting.js
│   │
│   ├── package.json              # Node dependencies
│   ├── tailwind.config.js        # Tailwind configuration
│   └── postcss.config.js         # PostCSS configuration
│
├── static/                       # Production build output
│   ├── index.html
│   ├── css/
│   ├── js/
│   └── assets/
│
├── docs/
│   ├── README.md                 # Main documentation
│   ├── PROJECT_STRUCTURE.md      # This file
│   ├── API.md                    # API documentation
│   └── CONFIGURATION.md          # Configuration guide
│
├── tests/                        # Test files (optional)
│   ├── backend/
│   │   ├── server_test.go
│   │   └── game_test.go
│   └── frontend/
│       └── App.test.js
│
├── scripts/                      # Utility scripts
│   ├── setup.sh                  # Initial setup script
│   ├── deploy.sh                 # Deployment script
│   └── reset-stats.sh            # Reset statistics
│
├── .gitignore                    # Git ignore patterns
├── .env.example                  # Environment variables template
├── Dockerfile                    # Docker configuration
├── docker-compose.yml            # Docker Compose setup
└── LICENSE                       # License file
```

## File Descriptions

### Backend Files

#### server.go
Main server application containing:
- WebSocket handler for real-time game communication
- HTTP endpoints for configuration, stats, and leaderboard
- AI provider integrations (OpenAI, Anthropic, Google, Ollama, HuggingFace)
- Game state management
- Scoring algorithm
- Statistics tracking
- Leaderboard management

Size: ~800 lines
Dependencies: gorilla/websocket

#### config.json
Configuration file defining:
- Active AI models
- Provider settings
- API keys
- Custom endpoints

Format: JSON
Required: Yes

#### stats.json
Auto-generated statistics file storing:
- Total games played
- Win/loss record
- Win rate
- Games by difficulty
- Average duration

Format: JSON
Generated: Automatically on first game

#### leaderboard.json
Auto-generated leaderboard file storing:
- Top 100 high scores
- Riddle details
- Game metadata
- Timestamps

Format: JSON
Generated: Automatically on first game

### Frontend Files

#### App.js
Main React component managing:
- Game state machine (menu, setup, playing, finished, stats, leaderboard)
- WebSocket communication
- UI rendering for all game screens
- Statistics display
- Leaderboard display

Size: ~600 lines
Dependencies: react, lucide-react

#### index.css
Tailwind CSS configuration:
- Base styles
- Component styles
- Utility classes

### Configuration Files

#### go.mod
Go module definition:
```go
module turing-roulette
go 1.21
require github.com/gorilla/websocket v1.5.1
```

#### package.json
Node.js dependencies:
```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "lucide-react": "^0.263.1"
  },
  "devDependencies": {
    "tailwindcss": "^3.3.0"
  }
}
```

#### tailwind.config.js
Tailwind CSS configuration:
```javascript
module.exports = {
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  theme: { extend: {} },
  plugins: [],
}
```

## Architecture

### Backend Architecture

```
┌─────────────────────────────────────────┐
│           HTTP Server (port 8080)        │
├─────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────────┐  │
│  │  WebSocket  │  │   HTTP Handlers  │  │
│  │   Handler   │  │  - /config       │  │
│  │             │  │  - /stats        │  │
│  │             │  │  - /leaderboard  │  │
│  └──────┬──────┘  └────────┬─────────┘  │
│         │                  │             │
│  ┌──────▼──────────────────▼─────────┐  │
│  │      Game State Manager           │  │
│  │  - Round management               │  │
│  │  - Model coordination             │  │
│  │  - Answer validation              │  │
│  └──────┬────────────────────────────┘  │
│         │                                │
│  ┌──────▼──────────────────────────┐    │
│  │   AI Provider Integrations      │    │
│  │  - OpenAI streaming             │    │
│  │  - Anthropic streaming          │    │
│  │  - Google API                   │    │
│  │  - Ollama local                 │    │
│  │  - HuggingFace inference        │    │
│  └──────┬──────────────────────────┘    │
│         │                                │
│  ┌──────▼──────────────────────────┐    │
│  │   Persistence Layer             │    │
│  │  - stats.json                   │    │
│  │  - leaderboard.json             │    │
│  │  - config.json                  │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

### Frontend Architecture

```
┌─────────────────────────────────────────┐
│         React Application                │
├─────────────────────────────────────────┤
│  ┌─────────────────────────────────┐    │
│  │      App Component (Root)        │    │
│  │  - State management              │    │
│  │  - Route handling                │    │
│  └────────┬─────────────────────────┘    │
│           │                               │
│  ┌────────▼──────────────────────────┐   │
│  │       Screen Components           │   │
│  │  - Menu                           │   │
│  │  - Game Setup                     │   │
│  │  - Game Playing                   │   │
│  │  - Game Results                   │   │
│  │  - Statistics                     │   │
│  │  - Leaderboard                    │   │
│  └────────┬──────────────────────────┘   │
│           │                               │
│  ┌────────▼──────────────────────────┐   │
│  │    Shared Components              │   │
│  │  - ModelColumn                    │   │
│  │  - DifficultyBadge               │   │
│  │  - ScoreCard                     │   │
│  └────────┬──────────────────────────┘   │
│           │                               │
│  ┌────────▼──────────────────────────┐   │
│  │      WebSocket Manager            │   │
│  │  - Connection handling            │   │
│  │  - Message parsing                │   │
│  │  - Streaming updates              │   │
│  └────────┬──────────────────────────┘   │
│           │                               │
│  ┌────────▼──────────────────────────┐   │
│  │       API Client                  │   │
│  │  - fetch /config                  │   │
│  │  - fetch /stats                   │   │
│  │  - fetch /leaderboard             │   │
│  └───────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

## Data Flow

### Game Initialization

```
User -> Setup Form -> WebSocket Send -> Server
                                         |
Server validates riddle -> Creates GameState
                                         |
Server -> Parallel goroutines for each model
                                         |
Each model streams response -> WebSocket -> Frontend
                                         |
Frontend updates UI in real-time
```

### Scoring Flow

```
Game ends -> Calculate base score
              |
              +-> Apply difficulty multiplier
              |
              +-> Add speed bonus
              |
              +-> Add stump bonus
              |
              v
          Final Score -> Update stats.json
                      -> Update leaderboard.json
                      -> Send to frontend
```

## State Management

### Backend State

```go
type GameState struct {
    Riddle       string
    Answer       string
    Clues        []string
    Difficulty   string
    CurrentRound int
    ModelStates  map[string]ModelState
    StartTime    time.Time
}
```

### Frontend State

```javascript
{
  gameState: 'menu|setup|playing|finished|stats|leaderboard',
  riddle: string,
  answer: string,
  clues: string[],
  difficulty: 'easy|medium|hard',
  modelOutputs: { [modelName]: string },
  modelResults: { [modelName]: boolean },
  currentRound: number,
  gameResult: GameResult | null,
  config: Config | null,
  stats: Stats | null,
  leaderboard: LeaderboardEntry[]
}
```

## Communication Protocols

### WebSocket Messages

#### Client to Server

```json
{
  "riddle": "string",
  "answer": "string",
  "clues": ["string", "string", "string"],
  "difficulty": "easy|medium|hard"
}
```

#### Server to Client

##### Streaming Guess
```json
{
  "type": "guess",
  "model": "string",
  "content": "string",
  "done": boolean
}
```

##### Result
```json
{
  "type": "result",
  "model": "string",
  "content": "true|false"
}
```

##### Game Result
```json
{
  "type": "gameResult",
  "correctCount": number,
  "totalModels": number,
  "gameOver": boolean,
  "playerWins": boolean,
  "duration": number,
  "score": number
}
```

### HTTP Endpoints

#### GET /config
Response:
```json
{
  "models": [
    {
      "name": "string",
      "provider": "string",
      "model": "string"
    }
  ]
}
```

#### GET /stats
Response:
```json
{
  "totalGames": number,
  "wins": number,
  "losses": number,
  "winRate": number,
  "byDifficulty": {
    "easy": number,
    "medium": number,
    "hard": number
  },
  "averageDuration": number
}
```

#### GET /leaderboard
Response:
```json
[
  {
    "riddle": "string",
    "difficulty": "string",
    "playerWon": boolean,
    "correctCount": number,
    "totalModels": number,
    "duration": number,
    "timestamp": "ISO8601",
    "score": number
  }
]
```

## Extension Points

### Adding New AI Providers

1. Add provider constants and structures in server.go
2. Implement `stream[Provider]` function
3. Add case in `streamModelResponse` switch
4. Update frontend icon mapping
5. Document configuration in README.md

### Adding New Game Modes

1. Extend `RiddleSubmission` struct with mode field
2. Add mode-specific logic in `playRound`
3. Create new scoring algorithm variant
4. Add UI components for new mode
5. Update statistics tracking

### Adding Authentication

1. Implement user management system
2. Add JWT token handling
3. Update WebSocket authentication
4. Add user-specific stats and leaderboards
5. Implement access control

### Adding Multiplayer

1. Implement room management
2. Add player synchronization
3. Create shared leaderboards
4. Implement turn-based mechanics
5. Add chat functionality

## Performance Considerations

### Backend Optimizations

- Use goroutine pools to limit concurrent model requests
- Implement response caching for repeated riddles
- Add request rate limiting
- Use connection pooling for external APIs
- Implement graceful shutdown

### Frontend Optimizations

- Lazy load leaderboard data
- Implement virtual scrolling for long lists
- Cache API responses locally
- Use React.memo for expensive components
- Debounce user input

## Security Best Practices

### Backend Security

- Validate all user input
- Sanitize riddle and clue text
- Implement rate limiting per IP
- Use HTTPS in production
- Never log API keys
- Implement CORS properly
- Add request size limits

### Frontend Security

- Validate WebSocket messages
- Sanitize displayed content
- Use environment variables for API URLs
- Implement CSP headers
- Avoid storing sensitive data in localStorage

## Deployment Considerations

### Development

```bash
# Backend
go run server.go

# Frontend
npm start
```

### Production

```bash
# Build frontend
npm run build

# Copy to static directory
cp -r build/* static/

# Build backend
go build -o turing-roulette server.go

# Run
./turing-roulette
```

### Docker

```bash
docker build -t turing-roulette .
docker run -p 8080:8080 turing-roulette
```

### Cloud Platforms

- AWS: Deploy to EC2, ECS, or Lambda
- GCP: Deploy to Compute Engine or Cloud Run
- Azure: Deploy to App Service or Container Instances
- Heroku: Use buildpacks for Go and static files

## Monitoring and Logging

### Recommended Logging

- Request/response for all API calls
- WebSocket connection events
- Model response times
- Error stack traces
- Game completion metrics

### Metrics to Track

- Average game duration
- Model accuracy rates
- API response times
- Error rates by provider
- User engagement metrics

## Future Enhancements

- Multi-language support
- Voice input for riddles
- Image-based riddles
- Team competitions
- Tournament mode
- Daily challenges
- Achievement system
- Social sharing
- Mobile app version
- AI difficulty adjustment