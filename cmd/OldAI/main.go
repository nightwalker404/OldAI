package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/nightwalker404/OldAI/internal/config"
	"github.com/nightwalker404/OldAI/internal/storage"
	"github.com/nightwalker404/OldAI/internal/vllmclient"
)

func main() {
	fmt.Println("Starting the application...")

	System := vllmclient.Message{
		Role:    "system",
		Content: "You are a helpful assistant.",
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	} else {
		log.Printf("Configuration loaded: %+v\n", cfg)
	}

	client := vllmclient.NewClient(cfg.VLLMModel, cfg.VLLMBaseURL)

	storage := storage.New(cfg.RedisHost+":"+strconv.Itoa(cfg.RedisPort), cfg.RedisPassword)
	log.Println("Storage initialized:", storage)

	history, err := storage.Load(cfg.RedisSession)
	if err != nil {
		log.Fatal("Failed to load history:", err)
	} else if len(history) == 0 {
		log.Println("No history found for session:", cfg.RedisSession)
		history = append(history, System)
	} else {
		log.Println("Loaded history for session completed")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("OldAI chat __ type 'exit' to quit")

	for {
		fmt.Print("\nYou: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		history = append(history, vllmclient.Message{Role: "user", Content: input})
		response, err := client.CreateRequest(history, vllmclient.ClientRequest{})

		if err != nil {
			log.Println("Error creating request:", err)
			continue
		}

		if len(response.Choices) > 0 {
			reply := response.Choices[0].Message.Content
			fmt.Println("OldAI:", reply)
			history = append(history, vllmclient.Message{Role: "assistant", Content: reply})
		} else {
			log.Println("No response from the model.")
		}
		storage.AppendMessage(cfg.RedisSession, vllmclient.Message{Role: "user", Content: input})
	}
}
