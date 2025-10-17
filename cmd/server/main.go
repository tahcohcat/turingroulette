package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Config struct {
	Models []ModelConfig `json:"models"`
}

type ModelConfig struct {
	Name     string `json:"name"`
	Provider string `json:"provider"` // "openai", "anthropic", "google", "ollama", "huggingface"
	Model    string `json:"model"`
	APIKey   string `json:"apiKey"`
	Endpoint string `json:"endpoint"`
}

type RiddleSubmission struct {
	Riddle     string   `json:"riddle"`
	Answer     string   `json:"answer"`
	Clues      []string `json:"clues"`
	Difficulty string   `json:"difficulty"` // "easy", "medium", "hard"
	Username   string   `json:"username"`
}

type GameState struct {
	Riddle         string                `json:"riddle"`
	Answer         string                `json:"answer"`
	Clues          []string              `json:"clues"`
	Difficulty     string                `json:"difficulty"`
	CurrentRound   int                   `json:"currentRound"`
	ModelStates    map[string]ModelState `json:"modelStates"`
	StartTime      time.Time             `json:"startTime"`
	Username       string                `json:"username"`
	SelectedModels []ModelConfig         `json:"selectedModels"`
}

type ModelState struct {
	Correct       bool      `json:"correct"`
	Guess         string    `json:"guess"`
	Round         int       `json:"round"` // Which round they got it correct
	AllGuesses    []string  `json:"allGuesses"` // History of all guesses
	GuessResults  []bool    `json:"guessResults"` // History of correct/incorrect for each guess
	ResponseTime  float64   `json:"responseTime"` // Response time in seconds
	ResponseTimes []float64 `json:"responseTimes"` // History of response times for each round
	GuessCount    int       `json:"guessCount"` // Track number of guesses made
	GuessesToCorrect int    `json:"guessesToCorrect"` // How many guesses needed to get correct
}

type StreamMessage struct {
	Model   string `json:"model"`
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Type    string `json:"type"` // "guess" or "result"
}

type GameResult struct {
	PlayerWins   bool      `json:"playerWins"`
	CorrectCount int       `json:"correctCount"`
	TotalModels  int       `json:"totalModels"`
	Difficulty   string    `json:"difficulty"`
	Duration     float64   `json:"duration"` // seconds
	RoundsPlayed int       `json:"roundsPlayed"`
	Timestamp    time.Time `json:"timestamp"`
	Username     string    `json:"username"`
}

type Stats struct {
	TotalGames      int                     `json:"totalGames"`
	Wins            int                     `json:"wins"`
	Losses          int                     `json:"losses"`
	WinRate         float64                 `json:"winRate"`
	ByDifficulty    map[string]int          `json:"byDifficulty"`
	AverageDuration float64                 `json:"averageDuration"`
	TotalDuration   float64                 `json:"totalDuration"`
	ByModel         map[string]ModelStats   `json:"byModel"`
}

type ModelStats struct {
	Name            string  `json:"name"`
	Provider        string  `json:"provider"`
	GamesPlayed     int     `json:"gamesPlayed"`
	TimesCorrect    int     `json:"timesCorrect"`
	Accuracy        float64 `json:"accuracy"`
	AvgResponseTime float64 `json:"avgResponseTime"`
	TotalResponseTime float64 `json:"totalResponseTime"`
	AvgGuessesToCorrect float64 `json:"avgGuessesToCorrect"`
	TotalGuessesToCorrect int   `json:"totalGuessesToCorrect"`
}

type LeaderboardEntry struct {
	Riddle       string                    `json:"riddle"`
	Difficulty   string                    `json:"difficulty"`
	Username     string                    `json:"username"`
	PlayerWon    bool                      `json:"playerWon"`
	CorrectCount int                       `json:"correctCount"`
	TotalModels  int                       `json:"totalModels"`
	Duration     float64                   `json:"duration"`
	Timestamp    time.Time                 `json:"timestamp"`
	Score        int                       `json:"score"` // Calculated score
	Models       []LeaderboardModelEntry   `json:"models"`
}

type LeaderboardModelEntry struct {
	Name          string  `json:"name"`
	Provider      string  `json:"provider"`
	Correct       bool    `json:"correct"`
	ResponseTime  float64 `json:"responseTime"`
	FinalGuess    string  `json:"finalGuess"`
}

// OpenAI structures
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// Anthropic structures
type AnthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream"`
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicStreamResponse struct {
	Type  string `json:"type"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

// Google Gemini structures
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []GeminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Ollama structures
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaStreamResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// HuggingFace structures
type HuggingFaceRequest struct {
	Inputs     string                 `json:"inputs"`
	Parameters HuggingFaceParameters  `json:"parameters"`
	Options    HuggingFaceOptions     `json:"options"`
}

type HuggingFaceParameters struct {
	MaxNewTokens int     `json:"max_new_tokens"`
	Temperature  float64 `json:"temperature"`
}

type HuggingFaceOptions struct {
	UseCache     bool `json:"use_cache"`
	WaitForModel bool `json:"wait_for_model"`
}

type HuggingFaceResponse struct {
	GeneratedText string `json:"generated_text"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var games = make(map[*websocket.Conn]*GameState)
var gamesMux sync.Mutex
var config Config
var stats Stats
var statsMux sync.Mutex
var leaderboard []LeaderboardEntry
var leaderboardMux sync.Mutex

const MAX_GUESSES = 3

var dataDir string

func init() {
	dataDir = os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data/"
	}
}

func main() {
	os.MkdirAll(dataDir, 0755)
	loadConfig()
	loadStats()
	loadLeaderboard()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleWebSocket)
	mux.HandleFunc("/config", handleGetConfig)
	mux.HandleFunc("/stats", handleGetStats)
	mux.HandleFunc("/leaderboard", handleGetLeaderboard)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Wrap the mux with the CORS middleware
	handler := corsMiddleware(mux)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// corsMiddleware allows local React dev (http://localhost:3000) to call your API
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from React dev server
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight (OPTIONS) requests quickly
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loadConfig() {
	file, err := os.ReadFile(dataDir + "config.json")
	if err != nil {
		log.Println("No config.json found, using default configuration")
		config = Config{
			Models: []ModelConfig{
				{Name: "Llama 2", Provider: "ollama", Model: "llama2", Endpoint: "http://localhost:11434"},
				{Name: "Mistral", Provider: "ollama", Model: "mistral", Endpoint: "http://localhost:11434"},
				{Name: "CodeLlama", Provider: "ollama", Model: "codellama", Endpoint: "http://localhost:11434"},
			},
		}
		return
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Error parsing config.json:", err)
	}

	// Override API keys with environment variables if they exist
	for i := range config.Models {
		envKey := fmt.Sprintf("%s_API_KEY", strings.ToUpper(config.Models[i].Provider))
		if envValue := os.Getenv(envKey); envValue != "" {
			config.Models[i].APIKey = envValue
		}
		// Also check for provider-specific env vars
		switch config.Models[i].Provider {
		case "openai":
			if key := os.Getenv("OPENAI_API_KEY"); key != "" {
				config.Models[i].APIKey = key
			}
		case "anthropic":
			if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
				config.Models[i].APIKey = key
			}
		case "google":
			if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
				config.Models[i].APIKey = key
			}
		case "huggingface":
			if key := os.Getenv("HUGGINGFACE_API_KEY"); key != "" {
				config.Models[i].APIKey = key
			}
		}
	}

	log.Printf("Loaded configuration with %d models\n", len(config.Models))
}

func loadStats() {
	file, err := os.ReadFile(dataDir + "stats.json")
	if err != nil {
		stats = Stats{
			ByDifficulty: make(map[string]int),
			ByModel:      make(map[string]ModelStats),
		}
		return
	}

	json.Unmarshal(file, &stats)
	if stats.ByModel == nil {
		stats.ByModel = make(map[string]ModelStats)
	}
}

func saveStats() {
	data, _ := json.MarshalIndent(stats, "", "  ")
	os.WriteFile(dataDir + "stats.json", data, 0644)
}

func loadLeaderboard() {
	file, err := os.ReadFile(dataDir + "leaderboard.json")
	if err != nil {
		leaderboard = []LeaderboardEntry{}
		return
	}

	json.Unmarshal(file, &leaderboard)
}

func saveLeaderboard() {
	data, _ := json.MarshalIndent(leaderboard, "", "  ")
	os.WriteFile(dataDir + "leaderboard.json", data, 0644)
}

func calculateScore(result GameResult) int {
	if !result.PlayerWins {
		return 0
	}

	baseScore := 100

	// Difficulty multiplier
	difficultyMultiplier := map[string]float64{
		"easy":   1.0,
		"medium": 1.5,
		"hard":   2.0,
	}

	multiplier := difficultyMultiplier[result.Difficulty]
	if multiplier == 0 {
		multiplier = 1.0
	}

	// Bonus for speed (max 50 points)
	timeBonus := 50.0
	if result.Duration > 60 {
		timeBonus = 50.0 * (60.0 / result.Duration)
	}

	// Bonus for stumping more models
	stumpBonus := float64((result.TotalModels - result.CorrectCount) * 20)

	score := float64(baseScore)*multiplier + timeBonus + stumpBonus
	return int(score)
}

func updateStats(result GameResult) {

log.Println("Updating stats with result:", result)
statsMux.Lock()
defer statsMux.Unlock()

stats.TotalGames++
if result.PlayerWins {
stats.Wins++
} else {
stats.Losses++
}

if stats.TotalGames > 0 {
stats.WinRate = float64(stats.Wins) / float64(stats.TotalGames) * 100
}

if stats.ByDifficulty == nil {
stats.ByDifficulty = make(map[string]int)
}
stats.ByDifficulty[result.Difficulty]++

stats.TotalDuration += result.Duration
stats.AverageDuration = stats.TotalDuration / float64(stats.TotalGames)

log.Println("Saving stats")
saveStats()
}

func updateModelStats(game *GameState) {
	statsMux.Lock()
	defer statsMux.Unlock()

	for _, modelCfg := range game.SelectedModels {
		if state, exists := game.ModelStates[modelCfg.Name]; exists {
			modelKey := modelCfg.Name

			if stats.ByModel == nil {
				stats.ByModel = make(map[string]ModelStats)
			}

			modelStat := stats.ByModel[modelKey]
			if modelStat.Name == "" {
				// Initialize new model stats
				modelStat = ModelStats{
					Name:     modelCfg.Name,
					Provider: modelCfg.Provider,
				}
			}

			modelStat.GamesPlayed++
			if state.Correct {
				modelStat.TimesCorrect++
				modelStat.TotalGuessesToCorrect += state.GuessesToCorrect
			}
			modelStat.TotalResponseTime += state.ResponseTime

			if modelStat.GamesPlayed > 0 {
				modelStat.Accuracy = float64(modelStat.TimesCorrect) / float64(modelStat.GamesPlayed) * 100
				modelStat.AvgResponseTime = modelStat.TotalResponseTime / float64(modelStat.GamesPlayed)
			}
			if modelStat.TimesCorrect > 0 {
				modelStat.AvgGuessesToCorrect = float64(modelStat.TotalGuessesToCorrect) / float64(modelStat.TimesCorrect)
			}

			stats.ByModel[modelKey] = modelStat
		}
	}

	saveStats()
}

func addToLeaderboard(game *GameState, result GameResult) {
	// Build model details for leaderboard
	var models []LeaderboardModelEntry
	for _, modelCfg := range game.SelectedModels {
		if state, exists := game.ModelStates[modelCfg.Name]; exists {
			// Get the final guess (last non-empty guess)
			finalGuess := ""
			if len(state.AllGuesses) > 0 {
				for i := len(state.AllGuesses) - 1; i >= 0; i-- {
					if state.AllGuesses[i] != "" {
						finalGuess = state.AllGuesses[i]
						break
					}
				}
			}

			models = append(models, LeaderboardModelEntry{
				Name:         modelCfg.Name,
				Provider:     modelCfg.Provider,
				Correct:      state.Correct,
				ResponseTime: state.ResponseTime,
				FinalGuess:   finalGuess,
			})
		}
	}

	entry := LeaderboardEntry{
		Riddle:       game.Riddle,
		Difficulty:   game.Difficulty,
		Username:     game.Username,
		PlayerWon:    result.PlayerWins,
		CorrectCount: result.CorrectCount,
		TotalModels:  result.TotalModels,
		Duration:     result.Duration,
		Timestamp:    result.Timestamp,
		Score:        calculateScore(result),
		Models:       models,
	}

	leaderboardMux.Lock()
	defer leaderboardMux.Unlock()

	leaderboard = append(leaderboard, entry)

	// Sort by score descending
	for i := 0; i < len(leaderboard)-1; i++ {
		for j := i + 1; j < len(leaderboard); j++ {
			if leaderboard[j].Score > leaderboard[i].Score {
				leaderboard[i], leaderboard[j] = leaderboard[j], leaderboard[i]
			}
		}
	}

	// Keep top 100
	if len(leaderboard) > 100 {
		leaderboard = leaderboard[:100]
	}

	saveLeaderboard()
}

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func handleGetStats(w http.ResponseWriter, r *http.Request) {
	statsMux.Lock()
	defer statsMux.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	leaderboardMux.Lock()
	defer leaderboardMux.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		var submission RiddleSubmission
		err := conn.ReadJSON(&submission)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		gamesMux.Lock()

		// Randomly select 3 models from config (or all if fewer than 3)
		selectedModels := config.Models
		if len(config.Models) > 3 {
			// Shuffle the models and take first 3
			shuffled := make([]ModelConfig, len(config.Models))
			copy(shuffled, config.Models)
			rand.Shuffle(len(shuffled), func(i, j int) {
				shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
			})
			selectedModels = shuffled[:3]
		}

		modelStates := make(map[string]ModelState)
		for _, model := range selectedModels {
			modelStates[model.Name] = ModelState{GuessCount: 0}
		}

		game := &GameState{
			Riddle:       submission.Riddle,
			Answer:       submission.Answer,
			Clues:        submission.Clues,
			Difficulty:   submission.Difficulty,
			CurrentRound: 0,
			ModelStates:  modelStates,
			StartTime:    time.Now(),
			Username:     submission.Username,
			SelectedModels: selectedModels,
		}
		games[conn] = game
		gamesMux.Unlock()

		// Send game start message with selected models
		startMsg := map[string]interface{}{
			"type":          "gameStart",
			"selectedModels": selectedModels,
		}
		conn.WriteJSON(startMsg)

		playRound(conn, game)
	}

	gamesMux.Lock()
	delete(games, conn)
	gamesMux.Unlock()
}

// Add this debugging code to cmd/server/main.go in the playRound function
// Right after checking results, add:

func playRound(conn *websocket.Conn, game *GameState) {
	// Send round start message
	conn.WriteJSON(map[string]interface{}{
		"type":  "roundStart",
		"round": game.CurrentRound,
	})

	var wg sync.WaitGroup
	for _, modelCfg := range game.SelectedModels {
		// Skip models that are already correct
		if game.ModelStates[modelCfg.Name].Correct {
			continue
		}

		wg.Add(1)
		go func(cfg ModelConfig) {
			defer wg.Done()
			prompt := buildPrompt(game, cfg.Name)
			streamModelResponse(conn, cfg, prompt, game)
		}(modelCfg)
	}

	wg.Wait()

	// Check results
	correctCount := 0
	for m, state := range game.ModelStates {
		if state.Correct {
			log.Printf("Model %s guessed correctly: %v\n", m, state.Guess)
			correctCount++
		}
	}

	totalModels := len(game.SelectedModels)
	allCorrect := correctCount == totalModels
	someCorrect := correctCount > 0 && correctCount < totalModels
	noneCorrect := correctCount == 0
	cluesExhausted := game.CurrentRound >= len(game.Clues)

	// ADD DEBUGGING HERE
	log.Printf("=== ROUND %d DEBUG ===\n", game.CurrentRound)
	log.Printf("Total Models: %d\n", totalModels)
	log.Printf("Correct Count: %d\n", correctCount)
	log.Printf("All Correct: %v\n", allCorrect)
	log.Printf("Some Correct: %v\n", someCorrect)
	log.Printf("None Correct: %v\n", noneCorrect)
	log.Printf("Clues Exhausted: %v (Round %d, Clues %d)\n", cluesExhausted, game.CurrentRound, len(game.Clues))
	log.Printf("Model States:\n")
	for name, state := range game.ModelStates {
		log.Printf("  %s: Correct=%v, Round=%d, Guess=%s\n", name, state.Correct, state.Round, state.Guess)
	}
	log.Printf("==================\n")

	result := map[string]interface{}{
		"type":           "gameResult",
		"correctCount":   correctCount,
		"totalModels":    totalModels,
		"allCorrect":     allCorrect,
		"someCorrect":    someCorrect,
		"cluesExhausted": cluesExhausted,
		"modelStates":    game.ModelStates,
	}

	// Game ends if all models correct OR all clues exhausted
	if allCorrect || cluesExhausted {
		log.Printf("GAME ENDING: allCorrect=%v, someCorrect=%v, cluesExhausted=%v", allCorrect, someCorrect, cluesExhausted)
		duration := time.Since(game.StartTime).Seconds()

		gameResult := GameResult{
			PlayerWins:   correctCount > 0 && correctCount < totalModels, // Win if some (but not all) models got correct
			CorrectCount: correctCount,
			TotalModels:  totalModels,
			Difficulty:   game.Difficulty,
			Duration:     duration,
			RoundsPlayed: game.CurrentRound + 1,
			Timestamp:    time.Now(),
			Username:     game.Username,
		}

		log.Printf("GAME FINISHED - Player Wins: %v\n", gameResult.PlayerWins)


		// Send game finished message with all result data
		finishedMsg := map[string]interface{}{
			"type":         "gameFinished",
			"playerWins":   gameResult.PlayerWins,
			"correctCount": correctCount,
			"totalModels":  totalModels,
			"duration":     duration,
			"score":        calculateScore(gameResult),
			"modelStates":  game.ModelStates,
		}

		// Add result message
		if gameResult.PlayerWins {
			finishedMsg["message"] = "🎉 You Win! Some AI guessed correctly, but not all."
		} else {
			if allCorrect {
				finishedMsg["message"] = "🤖 AI Wins! All AI guessed correctly."
			} else {
				finishedMsg["message"] = "🤖 AI Wins! No AI guessed correctly within the clues."
			}
		}

		log.Println("Sending gameFinished message")
		// Small delay so users can see the final results
		time.Sleep(2 * time.Second)
		conn.WriteJSON(finishedMsg)
		
		log.Println("Updating stats and leaderboard")
		updateStats(gameResult)
		updateModelStats(game)
	addToLeaderboard(game, gameResult)

		result["gameOver"] = true
		log.Print("Stats and leaderboard updated")

		// Pause before ending
		time.Sleep(1500 * time.Millisecond)

		return // End the game, don't continue
	} else {
		result["gameOver"] = false
		game.CurrentRound++
		result["nextRound"] = game.CurrentRound
	}

	conn.WriteJSON(result)

	time.Sleep(1500 * time.Millisecond)
	playRound(conn, game)
}

func buildPrompt(game *GameState, modelName string) string {
	prompt := fmt.Sprintf("Answer this riddle with just the answer (one or two words maximum):\n\n%s", game.Riddle)

	if game.CurrentRound > 0 && game.CurrentRound <= len(game.Clues) {
		cluesGiven := strings.Join(game.Clues[:game.CurrentRound], "\n")
		prompt = fmt.Sprintf("%s\n\nClues:\n%s\n\nProvide only the answer.", prompt, cluesGiven)
	}

	// Add history of incorrect guesses for this model
	state := game.ModelStates[modelName]
	var incorrectGuesses []string
	for i, guess := range state.AllGuesses {
		if !state.GuessResults[i] && strings.TrimSpace(guess) != "" {
			incorrectGuesses = append(incorrectGuesses, guess)
		}
	}
	if len(incorrectGuesses) > 0 {
		prompt += fmt.Sprintf("\n\nDo not repeat these previous incorrect guesses: %s", strings.Join(incorrectGuesses, ", "))
	}

	return prompt
}

func streamModelResponse(conn *websocket.Conn, modelCfg ModelConfig, prompt string, game *GameState) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var response string
	var err error

	switch modelCfg.Provider {
	case "openai":
		response, err = streamOpenAI(ctx, conn, modelCfg, prompt)
	case "anthropic":
		response, err = streamAnthropic(ctx, conn, modelCfg, prompt)
	case "google":
		response, err = streamGoogle(ctx, conn, modelCfg, prompt)
	case "ollama":
		response, err = streamOllama(ctx, conn, modelCfg, prompt)
	case "huggingface":
		response, err = streamHuggingFace(ctx, conn, modelCfg, prompt)
	default:
		err = fmt.Errorf("unknown provider: %s", modelCfg.Provider)
	}

	responseTime := time.Since(startTime).Seconds()

	// Trim and validate response
	response = strings.TrimSpace(response)

	var isCorrect bool
	if err != nil || response == "" {
		log.Printf("Error streaming from %s: %v\n", modelCfg.Name, err)
		isCorrect = false
		response = ""
	} else {
		isCorrect = checkAnswer(response, game.Answer)
	}

	gamesMux.Lock()
	state := game.ModelStates[modelCfg.Name]
	state.Guess = response
	state.GuessCount++
	state.ResponseTime = responseTime

	if isCorrect && !state.Correct {
		state.Correct = true
		state.Round = game.CurrentRound + 1
		state.GuessesToCorrect = state.GuessCount
	}

	// Add to history only if response is not empty
	if response != "" {
		state.AllGuesses = append(state.AllGuesses, response)
		state.GuessResults = append(state.GuessResults, isCorrect)
		state.ResponseTimes = append(state.ResponseTimes, responseTime)
	}

	game.ModelStates[modelCfg.Name] = state
	gamesMux.Unlock()

	// Only send result if no error (successful response)
	if err == nil && response != "" {
		resultMsg := StreamMessage{
			Model:   modelCfg.Name,
			Content: fmt.Sprintf("%v", isCorrect),
			Done:    true,
			Type:    "result",
		}
		conn.WriteJSON(resultMsg)
	}
}

func streamOpenAI(ctx context.Context, conn *websocket.Conn, cfg ModelConfig, prompt string) (string, error) {
	reqBody := OpenAIRequest{
		Model: cfg.Model,
		Messages: []OpenAIMessage{
			{Role: "user", Content: prompt},
		},
		Stream: true,
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var fullResponse strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var streamResp OpenAIStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) > 0 {
			content := streamResp.Choices[0].Delta.Content
			fullResponse.WriteString(content)

			msg := StreamMessage{
				Model:   cfg.Name,
				Content: content,
				Done:    false,
				Type:    "guess",
			}
			conn.WriteJSON(msg)
		}
	}

	return fullResponse.String(), nil
}

func streamAnthropic(ctx context.Context, conn *websocket.Conn, cfg ModelConfig, prompt string) (string, error) {
	reqBody := AnthropicRequest{
		Model: cfg.Model,
		Messages: []AnthropicMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens: 1024,
		Stream:    true,
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var fullResponse strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var streamResp AnthropicStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}

		if streamResp.Type == "content_block_delta" && streamResp.Delta.Type == "text_delta" {
			content := streamResp.Delta.Text
			fullResponse.WriteString(content)

			msg := StreamMessage{
				Model:   cfg.Name,
				Content: content,
				Done:    false,
				Type:    "guess",
			}
			conn.WriteJSON(msg)
		}
	}

	return fullResponse.String(), nil
}

func streamGoogle(ctx context.Context, conn *websocket.Conn, cfg ModelConfig, prompt string) (string, error) {
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models/%s:generateContent?key=%s", cfg.Model, cfg.APIKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		content := geminiResp.Candidates[0].Content.Parts[0].Text

		for _, char := range content {
			msg := StreamMessage{
				Model:   cfg.Name,
				Content: string(char),
				Done:    false,
				Type:    "guess",
			}
			conn.WriteJSON(msg)
			time.Sleep(20 * time.Millisecond)
		}

		return content, nil
	}

	return "", fmt.Errorf("no response from Gemini")
}

func streamOllama(ctx context.Context, conn *websocket.Conn, cfg ModelConfig, prompt string) (string, error) {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	reqBody := OllamaRequest{
		Model:  cfg.Model,
		Prompt: prompt,
		Stream: true,
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var streamResp OllamaStreamResponse
		if err := decoder.Decode(&streamResp); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		fullResponse.WriteString(streamResp.Response)

		msg := StreamMessage{
			Model:   cfg.Name,
			Content: streamResp.Response,
			Done:    streamResp.Done,
			Type:    "guess",
		}
		conn.WriteJSON(msg)

		if streamResp.Done {
			break
		}
	}

	return fullResponse.String(), nil
}

func streamHuggingFace(ctx context.Context, conn *websocket.Conn, cfg ModelConfig, prompt string) (string, error) {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://api-inference.huggingface.co/models/%s", cfg.Model)
	}

	reqBody := HuggingFaceRequest{
		Inputs: prompt,
		Parameters: HuggingFaceParameters{
			MaxNewTokens: 100,
			Temperature:  0.7,
		},
		Options: HuggingFaceOptions{
			UseCache:     false,
			WaitForModel: true,
		},
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var hfResp []HuggingFaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&hfResp); err != nil {
		return "", err
	}

	if len(hfResp) > 0 {
		content := hfResp[0].GeneratedText

		// Remove the prompt from the response if it's included
		content = strings.TrimPrefix(content, prompt)
		content = strings.TrimSpace(content)

		// Simulate streaming
		for _, char := range content {
			msg := StreamMessage{
				Model:   cfg.Name,
				Content: string(char),
				Done:    false,
				Type:    "guess",
			}
			conn.WriteJSON(msg)
			time.Sleep(20 * time.Millisecond)
		}

		return content, nil
	}

	return "", fmt.Errorf("no response from HuggingFace")
}

func checkAnswer(guess string, correctAnswer string) bool {
	guess = strings.TrimSpace(strings.ToLower(guess))
	answer := strings.TrimSpace(strings.ToLower(correctAnswer))

	guess = strings.TrimPrefix(guess, "the answer is ")
	guess = strings.TrimPrefix(guess, "i believe the answer is ")
	guess = strings.TrimPrefix(guess, "based on the clues, it's ")
	guess = strings.TrimPrefix(guess, "it's ")
	guess = strings.TrimPrefix(guess, "a ")
	guess = strings.TrimPrefix(guess, "an ")
	guess = strings.TrimSuffix(guess, "?")
	guess = strings.TrimSuffix(guess, ".")

	return strings.Contains(guess, answer) || strings.Contains(answer, guess) || guess == answer
}