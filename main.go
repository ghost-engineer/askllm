package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Message describes a single message for DeepSeek API
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekRequestPayload represents the request structure for Chutes DeepSeek API
type DeepSeekRequestPayload struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

// Choice describes a single response option from DeepSeek API
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// UsageInfo contains token usage information
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// DeepSeekResponsePayload represents the response structure from DeepSeek API
type DeepSeekResponsePayload struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   UsageInfo `json:"usage"`
}

func main() {
	// Get Chutes API token from environment variable
	apiKey := os.Getenv("CHUTES_API_TOKEN")
	if apiKey == "" {
		log.Fatal("Error: CHUTES_API_TOKEN environment variable is not set.")
	}

	// Initialize Gin
	router := gin.Default()

	// Define route for root URL
	router.GET("/", func(c *gin.Context) {
		// Get 'q' parameter from URL query (user's prompt)
		query := c.Query("q")

		if query == "" {
			c.String(http.StatusBadRequest, "Please provide a query with the 'q' parameter. Example: /?q=Hello")
			return
		}

		log.Printf("Received request for DeepSeek: %s", query)

		// Build payload for Chutes DeepSeek API request
		payload := DeepSeekRequestPayload{
			Model:       "deepseek-ai/DeepSeek-R1",
			Messages:    []Message{{Role: "user", Content: query}},
			Stream:      false,
			MaxTokens:   1024,
			Temperature: 0.7,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshaling JSON request for DeepSeek: %v", err)
			c.String(http.StatusInternalServerError, "Internal server error.")
			return
		}

		// URL for Chutes DeepSeek API
		apiUrl := "https://llm.chutes.ai/v1/chat/completions"

		// Create HTTP client
		client := &http.Client{
			Timeout: 60 * time.Second, // Increase timeout if LLM may respond slowly
		}

		// Create HTTP request
		req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonPayload))
		if err != nil {
			log.Printf("Error creating HTTP request for DeepSeek: %v", err)
			c.String(http.StatusInternalServerError, "Internal server error.")
			return
		}
		req.Header.Set("Content-Type", "application/json")
		// Add Authorization header with your API key
		req.Header.Set("Authorization", "Bearer "+apiKey)

		// Send request to DeepSeek API
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending request to DeepSeek API: %v", err)
			c.String(http.StatusInternalServerError, "Failed to contact DeepSeek LLM. Please try again later.")
			return
		}
		defer resp.Body.Close() // Close response body after use

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body from DeepSeek API: %v", err)
			c.String(http.StatusInternalServerError, "Internal server error.")
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error from DeepSeek API. Status: %d, Body: %s", resp.StatusCode, string(body))
			c.String(http.StatusInternalServerError, "Error from DeepSeek LLM. Please try again later.")
			return
		}

		// Decode JSON response from DeepSeek API
		var deepseekResponse DeepSeekResponsePayload
		err = json.Unmarshal(body, &deepseekResponse)
		if err != nil {
			log.Printf("Error decoding JSON response from DeepSeek API: %v", err)
			c.String(http.StatusInternalServerError, "Internal server error: invalid response format from DeepSeek LLM.")
			return
		}

		// Extract response text
		if len(deepseekResponse.Choices) > 0 && deepseekResponse.Choices[0].Message.Content != "" {
			llmText := deepseekResponse.Choices[0].Message.Content
			log.Printf("DeepSeek LLM response: %s", llmText)
			c.String(http.StatusOK, llmText) // Send plain response text to user
		} else {
			log.Println("DeepSeek LLM did not provide a text response.")
			c.String(http.StatusOK, "DeepSeek LLM could not generate a response to your query.")
		}
	})

	// Start server on port 8080
	log.Println("AskLLM.io (DeepSeek) server started on port :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
