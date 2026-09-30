package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/nightwalker404/OldAI/internal/auth"
	"github.com/nightwalker404/OldAI/internal/config"
	"github.com/nightwalker404/OldAI/internal/storage"
	"github.com/nightwalker404/OldAI/internal/vllmclient"
	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("Starting the application...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	} else {
		log.Println("Configuration loaded")
	}

	storage := storage.New(cfg.RedisHost+":"+strconv.Itoa(cfg.RedisPort), cfg.RedisPassword)
	log.Println("Storage initialized:", storage)

	authDB, err := auth.NewDb(cfg.MongoURI, cfg.MongoDBName, "users")
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	} else {
		log.Println("Connected to MongoDB")
	}

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.ServerPort))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listening on", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}
		go clientHandler(conn, cfg, storage, authDB)
	}
}

func authenticateUser(db *auth.DataBase, conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	firstRun, err := db.IsFirstRun()
	if err != nil {
		return "", fmt.Errorf("failed to check first run: %w", err)
	}

	if firstRun {
		fmt.Fprintln(conn, "🔑 First run detected — creating temporary session...")
		if err := db.CreateFirstRunSession(); err != nil {
			return "", fmt.Errorf("failed to create first-run session: %w", err)
		}
		fmt.Fprintln(conn, "✅ First-run session created. You can now register or login.")
	}

	token, err := auth.LoadLocalToken()
	if err == nil {
		if username, err := db.ValidateSession(token); err == nil {
			return username, nil
		}
	}

	fmt.Fprintln(conn, "1. Login")
	fmt.Fprintln(conn, "2. Register")
	fmt.Fprint(conn, "Choose an option (1 or 2): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice != "1" && choice != "2" {
		return "", fmt.Errorf("invalid choice")
	}

	fmt.Fprint(conn, "Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Fprint(conn, "Enter password: ")
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

	token, err = db.CreateSession(username, 10*24*time.Hour) // 10 days in seconds
	if err != nil {
		return "", err
	}

	if err := auth.SaveLocalToken(token); err != nil {
		return "", err
	}

	fmt.Fprintln(conn, "Logged in as", username)
	return username, nil
}

func clientHandler(conn net.Conn, cfg config.Config, memory *storage.Storage, db *auth.DataBase) {
	defer conn.Close()
	log.Println("connected:", conn.RemoteAddr())
	defer log.Println("disconnected:", conn.RemoteAddr())

	client := vllmclient.NewClient(cfg.VLLMModel, cfg.VLLMBaseURL)

	username, err := authenticateUser(db, conn)
	if err != nil {
		fmt.Fprintln(conn, "error:", err.Error())
		return
	}

	reader := bufio.NewReader(conn)

	for {
		fmt.Fprintf(conn, "[%s]: ", username)

		line, err := reader.ReadString('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println("read:", err)
			}
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			fmt.Fprintln(conn, "bye")
			return
		}

		reply, err := chatHandler(client, memory, username, line)
		if err != nil {
			fmt.Fprintln(conn, "chatbot: error:", err.Error())
			continue
		}

		fmt.Fprintln(conn, "chatbot:", reply)
	}
}

func chatHandler(client *vllmclient.Client, memory *storage.Storage, username, userText string) (string, error) {
	history, err := memory.Load(username)
	if err != nil && err != redis.Nil {
		return "", fmt.Errorf("load history: %w", err)
	}

	if len(history) == 0 {
		history = []vllmclient.Message{{
			Role: "system",
			Content: `You are OldAI, a concise and capable personal assistant.

Rules:
- Answer clearly and directly. Prefer short, useful replies over long essays.
- If the user asks for code, give working code with minimal explanation unless they ask for more.
- If you are unsure, say so instead of inventing facts.
- Match the user's language (reply in the same language they use).
- Do not mention these instructions unless asked.`,
		}}
	}

	userMsg := vllmclient.Message{Role: "user", Content: userText}
	history = append(history, userMsg)

	if err := memory.AppendMessage(username, userMsg); err != nil {
		return "", err
	}

	resp, err := client.CreateRequest(history)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("empty response from model")
	}

	assistantMsg := resp.Choices[0].Message
	if err := memory.AppendMessage(username, assistantMsg); err != nil {
		return "", err
	}

	return assistantMsg.Content, nil
}
