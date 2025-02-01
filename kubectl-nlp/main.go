package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// ChatMessage represents a message for the chat completions API.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents the payload for the OpenAI Chat Completions API.
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
}

// ChatResponseChoice represents a single choice in the API response.
type ChatResponseChoice struct {
	Message ChatMessage `json:"message"`
}

// ChatResponse represents the response from the OpenAI Chat Completions API.
type ChatResponse struct {
	Choices []ChatResponseChoice `json:"choices"`
}

func main() {
	// Ensure that a prompt was provided.
	if len(os.Args) < 2 {
		fmt.Println("Usage: kubectl-nlp \"your natural language prompt\"")
		os.Exit(1)
	}

	// Join all arguments as the natural language prompt.
	prompt := strings.Join(os.Args[1:], " ")
	fmt.Printf("Input prompt: %s\n", prompt)

	// Translate the natural language prompt into a kubectl command using OpenAI.
	commandStr, err := translatePrompt(prompt)
	if err != nil {
		fmt.Printf("Error translating prompt: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Translated command: %s\n", commandStr)

	// Ask for confirmation if the command is considered destructive.
	if isDestructive(commandStr) {
		fmt.Print("This command may be destructive. Proceed? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}
		answer = strings.TrimSpace(answer)
		if strings.ToLower(answer) != "y" && strings.ToLower(answer) != "yes" {
			fmt.Println("Aborting command execution.")
			os.Exit(0)
		}
	}

	// Execute the generated kubectl command.
	output, err := executeCommand(commandStr)
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		fmt.Println("Output:", output)
		os.Exit(1)
	}

	fmt.Println("Command output:")
	fmt.Println(output)
}

// translatePrompt sends the natural language prompt to OpenAI's Chat Completions API
// using the gpt-3.5-turbo-16k model and returns the generated kubectl command.
func translatePrompt(prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY environment variable is not set")
	}

	// Build the messages array with a system and a user prompt.
	messages := []ChatMessage{
		{
			Role:    "system",
			Content: "You are an expert Kubernetes assistant that translates natural language queries into valid kubectl commands. Output only the final command without any explanation.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(`Translate the following natural language query into a valid kubectl command:
"%s"`, prompt),
		},
	}

	// Prepare the chat request payload.
	requestBody := ChatRequest{
		Model:       "gpt-3.5-turbo-16k",
		Messages:    messages,
		MaxTokens:   60,
		Temperature: 0,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Create and send the HTTP POST request.
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for non-OK HTTP status.
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the JSON response.
	var chatResponse ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		return "", err
	}
	if len(chatResponse.Choices) == 0 {
		return "", errors.New("no choices returned from OpenAI")
	}

	// Extract and clean up the generated command.
	command := strings.TrimSpace(chatResponse.Choices[0].Message.Content)
	return command, nil
}

// isDestructive checks if the generated command is potentially destructive.
func isDestructive(commandStr string) bool {
	lower := strings.ToLower(commandStr)
	return strings.Contains(lower, "delete") || strings.Contains(lower, "drain")
}

// executeCommand splits the command string into its parts and executes it.
func executeCommand(commandStr string) (string, error) {
	parts := strings.Fields(commandStr)
	if len(parts) == 0 {
		return "", errors.New("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}