# OldAI

A self-hosted CLI chatbot. Register or log in, then chat with a locally-running LLM (served via vLLM) straight from your terminal — with persistent conversation history and account-based sessions.

## Features

- 🔐 **Accounts** — register/login with bcrypt-hashed passwords, backed by MongoDB
- 🪪 **Session persistence** — log in once; a local session token keeps you signed in on future runs
- 💬 **Persistent chat history** — conversations are stored per-user in Redis and reloaded automatically
- 🤖 **Local inference** — powered by [vLLM](https://github.com/vllm-project/vllm) serving an OpenAI-compatible API, running entirely on your own machine/GPU
- 🐳 **One-command infra** — Redis, MongoDB, and vLLM all run via Docker Compose

## Architecture

```
┌─────────────┐      ┌───────────────┐
│   OldAI CLI │──────▶│  MongoDB      │  accounts, sessions
│  (Go binary)│      └───────────────┘
│             │      ┌───────────────┐
│             │──────▶│  Redis        │  chat history
│             │      └───────────────┘
│             │      ┌───────────────┐
│             │──────▶│  vLLM         │  model inference
└─────────────┘      └───────────────┘
```

The CLI is the only thing you run outside Docker — Mongo, Redis, and vLLM are containerized dependencies.

## Prerequisites

- [Go](https://go.dev/) 1.21+
- [Docker](https://www.docker.com/) and Docker Compose
- An NVIDIA GPU with drivers + [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) installed (required for the vLLM container)

## Setup

**1. Clone the repo and create your `.env` file** in the project root:

```env
### Redis Configuration ###
RedisPort=6479
RedisHost=localhost
RedisSession=DefaultSession
RedisPassword=change-me

### VLLM Configuration ###
VLLM_PORT=8003
VLLM_BASE_URL=http://localhost:8003
VLLM_TIMEOUT=60s
VLLM_MODEL=Qwen/Qwen2.5-0.5B-Instruct
VLLM_GPU_MEMORY_UTILIZATION=0.8
VLLM_MAX_MODEL_LEN=1024

### MongoDB Configuration ###
MongoHost=localhost
MongoPort=27017
MongoUser=admin
MongoPassword=change-me
MONGO_DB_NAME=OldAI
MONGO_URI=mongodb://admin:change-me@localhost:27017
```

> Use different passwords for Redis and MongoDB, and never commit `.env` to version control.

**2. Start the infrastructure:**

```bash
docker compose up -d
```

Give vLLM a minute or two on first run — it needs to download and load the model. You can check readiness with:

```bash
docker compose ps
```

Wait until `OldAIVLLM` shows `healthy`.

**3. Build and run the CLI:**

```bash
go run cmd/OldAI/main.go
```

## Usage

On first run you'll be asked to **register** a new account. After that, logging in once keeps you signed in on future runs via a local session token (stored at `~/.oldai/session.json`).

```
1. Login
2. Register
Choose an option (1 or 2): 2
Enter username: alice
Enter password: ********

Welcome to OldAI chat, alice! Type 'exit' to quit.

You: hey, what can you help me with?
OldAI: I'm a helpful assistant — ask me anything!

You: exit
```

Your conversation is saved automatically and picked back up the next time you log in as the same user.