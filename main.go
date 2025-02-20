package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// OpenAI-compatible request structure
type OpenAIRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// OpenAI-compatible response structure
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Load a Hugging Face model (runs as a subprocess)
func loadModel(modelName string) error {
	fmt.Println("Downloading model:", modelName)

	token := os.Getenv("HUGGINGFACE_TOKEN")
	if token == "" {
		return fmt.Errorf("HUGGINGFACE_TOKEN is not set")
	}

	fmt.Println("Downloading model:", modelName)

	// Run the CLI command with authentication
	cmd := exec.Command("huggingface-cli", "download", modelName, "--local-dir", "./models/"+modelName)
	cmd.Env = append(os.Environ(), "HF_TOKEN="+token) // Pass token as an environment variable
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error downloading model:", string(output))
		return err
	}

	fmt.Println("Model downloaded successfully")
	return nil
}

// Generate text using Hugging Face model (calls `llama.cpp` or a similar tool)
func generateText(prompt string, modelName string) (string, error) {
	//cmd := exec.Command("llama-cli", "-m", "./models/"+modelName+"/mistral-7b-v0.1.Q2_K.gguf", "-p", prompt, "--temp", "0.7", "-n", "256")
	cmd := exec.Command("llama-cli", "-m", "./models/"+modelName+"/tiny-vicuna-1b.q2_k.gguf", "-p", prompt, "--temp", "0.7", "-n", "256")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error generating text:", string(output))
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func main() {
	router := gin.Default()

	// Endpoint to generate text
	router.POST("/v1/completions", func(c *gin.Context) {
		var request OpenAIRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		responseText, err := generateText(request.Prompt, request.Model)
		if err != nil {
			log.Printf("Error generating text: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate text"})
			return
		}

		response := OpenAIResponse{
			ID:      "chatcmpl-123",
			Object:  "text_completion",
			Created: 1234567890,
			Model:   request.Model,
			Choices: []struct {
				Text         string `json:"text"`
				Index        int    `json:"index"`
				FinishReason string `json:"finish_reason"`
			}{
				{Text: responseText, Index: 0, FinishReason: "stop"},
			},
		}

		c.JSON(http.StatusOK, response)
	})

	// Load model before starting API
	//modelName := "mistralai/Mistral-7B-Instruct-v0.1"
	//modelName := "TheBloke/Mistral-7B-v0.1-GGUF"
	modelName := "afrideva/Tiny-Vicuna-1B-GGUF"
	//modelName := "SakanaAI/Llama-3-8B-Instruct-Coding-Expert"
	err := loadModel(modelName)
	if err != nil {
		log.Fatal("Failed to load model:", err)
	}

	// Start the server
	fmt.Println("Starting OpenAI-compatible API on port 8000...")
	router.Run(":8000")
}

