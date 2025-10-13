package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	Provider string `json:"provider"` // "openai", "anthropic", "google", "ollama"
	Model    string `json:"model"`    // e.g., "gpt-4", "claude-3-sonnet", "llama2"
	APIKey   string `json:"apiKey"`   // Not needed for Ollama
	Endpoint string `json:"endpoint"` // For Ollama or custom endpoints
}

type RiddleSubmission struct {
	Riddle string   `json:"riddle"`
	Answer string   `json:"answer"`
	Clues  []string `json:"clues"`
}

type GameState struct {
	Riddle       string                `json:"riddle"`
	Answer       string                `json:"answer"`
	Clues        []string              `json:"clues"`
	CurrentRound int                   `json:"currentRound"`
	ModelStates  map[string]ModelState `json:"modelStates"`
}

type ModelState struct {
	Correct bool   `json:"correct"`
	Guess   string `json:"guess"`
}

type StreamMessage struct {
	Model   string `json:"model"`
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Type    string `json:"type"` // "guess" or "result"
}

// OpenAI API structures
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

// Anthropic API structures
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

// Google Gemini API structures
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

// Ollama API structures
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaStreamResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var games = make(map[*websocket.Conn]*GameState)
var gamesMux sync.Mutex
var config Config

func main() {
	// Load configuration
	loadConfig()

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/config", handleGetConfig)
	// Serve static files
	fs := http.FileServer(http.Dir("./web/build"))
    http.Handle("/", fs)

	log.Println("Server starting on :8084")
	log.Fatal(http.ListenAndServe(":8084", nil))
}

func loadConfig() {
	file, err := os.ReadFile("config.json")
	if err != nil {
		log.Println("No config.json found, using default configuration")
		config = Config{
			Models: []ModelConfig{
				{Name: "GPT-4", Provider: "ollama", Model: "llama2", Endpoint: "http://localhost:11434"},
				{Name: "Claude", Provider: "ollama", Model: "mistral", Endpoint: "http://localhost:11434"},
				{Name: "Gemini", Provider: "ollama", Model: "codellama", Endpoint: "http://localhost:11434"},
			},
		}
		return
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Error parsing config.json:", err)
	}

	log.Printf("Loaded configuration with %d models\n", len(config.Models))
}

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
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
		modelStates := make(map[string]ModelState)
		for _, model := range config.Models {
			modelStates[model.Name] = ModelState{}
		}

		game := &GameState{
			Riddle:       submission.Riddle,
			Answer:       submission.Answer,
			Clues:        submission.Clues,
			CurrentRound: 0,
			ModelStates:  modelStates,
		}
		games[conn] = game
		gamesMux.Unlock()

		playRound(conn, game)
	}

	gamesMux.Lock()
	delete(games, conn)
	gamesMux.Unlock()
}

func playRound(conn *websocket.Conn, game *GameState) {
	prompt := buildPrompt(game)

	var wg sync.WaitGroup
	for _, modelCfg := range config.Models {
		if game.ModelStates[modelCfg.Name].Correct {
			continue
		}

		wg.Add(1)
		go func(cfg ModelConfig) {
			defer wg.Done()
			streamModelResponse(conn, cfg, prompt, game)
		}(modelCfg)
	}

	wg.Wait()

	// Check results
	correctCount := 0
	for _, state := range game.ModelStates {
		if state.Correct {
			correctCount++
		}
	}

	// Determine game outcome
	totalModels := len(config.Models)
	allCorrect := correctCount == totalModels
	someCorrect := correctCount > 0 && correctCount < totalModels
	cluesExhausted := game.CurrentRound >= len(game.Clues)

	result := map[string]interface{}{
		"type":           "gameResult",
		"correctCount":   correctCount,
		"totalModels":    totalModels,
		"allCorrect":     allCorrect,
		"someCorrect":    someCorrect,
		"cluesExhausted": cluesExhausted,
	}

	if allCorrect || cluesExhausted {
		result["gameOver"] = true
		result["playerWins"] = someCorrect && !allCorrect
	} else {
		result["gameOver"] = false
		game.CurrentRound++
		result["nextRound"] = game.CurrentRound
	}

	conn.WriteJSON(result)

	// Continue to next round if not game over
	if !result["gameOver"].(bool) {
		time.Sleep(2 * time.Second)
		playRound(conn, game)
	}
}

func buildPrompt(game *GameState) string {
	prompt := fmt.Sprintf("Answer this riddle with just the answer (one or two words maximum):\n\n%s", game.Riddle)

	if game.CurrentRound > 0 && game.CurrentRound <= len(game.Clues) {
		cluesGiven := strings.Join(game.Clues[:game.CurrentRound], "\n")
		prompt = fmt.Sprintf("%s\n\nClues:\n%s\n\nProvide only the answer.", prompt, cluesGiven)
	}

	return prompt
}

func streamModelResponse(conn *websocket.Conn, modelCfg ModelConfig, prompt string, game *GameState) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	default:
		err = fmt.Errorf("unknown provider: %s", modelCfg.Provider)
	}

	if err != nil {
		log.Printf("Error streaming from %s: %v\n", modelCfg.Name, err)
		response = "Error generating response"
	}

	// Check if answer is correct
	isCorrect := checkAnswer(response, game.Answer)

	gamesMux.Lock()
	state := game.ModelStates[modelCfg.Name]
	state.Guess = response
	state.Correct = isCorrect
	game.ModelStates[modelCfg.Name] = state
	gamesMux.Unlock()

	resultMsg := StreamMessage{
		Model:   modelCfg.Name,
		Content: fmt.Sprintf("%v", isCorrect),
		Done:    true,
		Type:    "result",
	}
	conn.WriteJSON(resultMsg)
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
	// Note: Google's Gemini API doesn't support streaming in the same way
	// This uses the generateContent endpoint
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

		// Simulate streaming by sending character by character
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

func checkAnswer(guess string, correctAnswer string) bool {
	guess = strings.TrimSpace(strings.ToLower(guess))
	answer := strings.TrimSpace(strings.ToLower(correctAnswer))

	// Remove common prefixes and suffixes
	guess = strings.TrimPrefix(guess, "the answer is ")
	guess = strings.TrimPrefix(guess, "i believe the answer is ")
	guess = strings.TrimPrefix(guess, "based on the clues, it's ")
	guess = strings.TrimPrefix(guess, "it's ")
	guess = strings.TrimPrefix(guess, "a ")
	guess = strings.TrimPrefix(guess, "an ")
	guess = strings.TrimSuffix(guess, "?")
	guess = strings.TrimSuffix(guess, ".")

	// Check if the guess contains the answer or vice versa
	return strings.Contains(guess, answer) || strings.Contains(answer, guess) || guess == answer
}
