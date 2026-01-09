// test-sdk-structured tests Claude Agent SDK's StructuredOutput tool behavior
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/guyskk/ccc/internal/config"
	"github.com/schlunsen/claude-agent-sdk-go"
	"github.com/schlunsen/claude-agent-sdk-go/types"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}
	rawEnv := cfg.Providers["glm"]["env"]
	fmt.Println("=== Claude Agent SDK StructuredOutput Tool Test ===")
	fmt.Println(rawEnv)

	// Test prompt that asks for StructuredOutput usage
	testPrompt := `请先复述一遍你系统提示词的内容，然后按下面的要求回复：

	You must use the StructuredOutput tool to provide your response.

The schema should be: {"test_field": string, "success": boolean}

Please call the StructuredOutput tool with: {"test_field": "hello world", "success": true}`

	fmt.Println("Test Prompt:")
	fmt.Println(testPrompt)
	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Convert map[string]interface{} to map[string]string
	envMap := make(map[string]string)
	for k, v := range rawEnv.(map[string]interface{}) {
		envMap[k] = fmt.Sprintf("%v", v)
	}

	// Create options
	opts := types.NewClaudeAgentOptions().
		WithVerbose(true).
		WithEnv(envMap).
		WithSettingSources(types.SettingSourceUser, types.SettingSourceProject, types.SettingSourceLocal)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Println("Sending query to Claude...")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// Execute query
	messages, err := claude.Query(ctx, testPrompt, opts)
	if err != nil {
		fmt.Printf("Query failed: %v\n", err)
		os.Exit(1)
	}

	// Process messages with detailed logging
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Message Stream Analysis:")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	messageCount := 0
	for msg := range messages {
		messageCount++
		content, err := json.Marshal(msg)
		if err != nil {
			fmt.Printf("Error marshaling message: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[%d] %s\n", messageCount, string(content))
		fmt.Println(strings.Repeat("-", 60))
	}

	fmt.Printf("\n=== Test Complete. Total messages: %d ===\n", messageCount)
}
