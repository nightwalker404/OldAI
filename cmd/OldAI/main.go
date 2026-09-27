package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/nightwalker404/OldAI/internal/auth"
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
		log.Println("Configuration loaded")
	}

	client := vllmclient.NewClient(cfg.VLLMModel, cfg.VLLMBaseURL)

	storage := storage.New(cfg.RedisHost+":"+strconv.Itoa(cfg.RedisPort), cfg.RedisPassword)
	log.Println("Storage initialized:", storage)

	authDB, err := auth.NewDb(cfg.MongoURI, cfg.MongoDBName, "users")
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	} else {
		log.Println("Connected to MongoDB")
	}

	reader := bufio.NewReader(os.Stdin)
	username, err := authenticateUser(authDB, reader)
	if err != nil {
		log.Fatal("Authentication failed:", err)
	} else {
		log.Println("User authenticated:", username)
	}

	fmt.Println("Welcome to OldAI chat, " + username + "! Type 'exit' to quit.")

	history, err := storage.Load(username)
	if err != nil {
		log.Fatal("Failed to load history:", err)
	} else if len(history) == 0 {
		log.Println("No history found for session:", cfg.RedisSession)
		history = append(history, System)
	} else {
		log.Println("Loaded history for session completed")
	}

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

func authenticateUser(db *auth.DataBase, reader *bufio.Reader) (string, error) {
	firstRun, err := db.IsFirstRun()
	if err != nil {
		return "", fmt.Errorf("failed to check first run: %w", err)
	}

	if firstRun {
		fmt.Println("🔑 First run detected — creating temporary session...")
		if err := db.CreateFirstRunSession(); err != nil {
			return "", fmt.Errorf("failed to create first-run session: %w", err)
		}
		fmt.Println("✅ First-run session created. You can now register or login.")
	}

	token, err := auth.LoadLocalToken()
	if err != nil {
		return "", err
	}

	username, err := db.ValidateSession(token)
	if err == nil {
		return username, nil
	}

	fmt.Println("1. Login")
	fmt.Println("2. Register")
	fmt.Print("Choose an option (1 or 2): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice != "1" && choice != "2" {
		return "", fmt.Errorf("invalid choice")
	}

	fmt.Print("Enter username: ")
	username, _ = reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if choice == "1" {
		if err := db.Login(username, password); err != nil {
			return "", err
		}
	} else {
		if err := db.Register(username, password); err != nil {
			return "", err
		}
	}

	token, err = db.CreateSession(username, 10*24*60*60) // 10 days in seconds
	if err != nil {
		return "", err
	}

	if err := auth.SaveLocalToken(token); err != nil {
		return "", err
	}

	return username, nil
}
