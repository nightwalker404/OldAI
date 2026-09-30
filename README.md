# OldAI

A self-hosted TCP chatbot server. Connect from your terminal (e.g. with `nc`), register or log in, then chat with a locally-running LLM (served via vLLM) — with persistent conversation history and account-based sessions.

## Features

- 🔐 **Accounts** — register/login with bcrypt-hashed passwords, backed by MongoDB
- 🪪 **Sessions** — login creates a session token (stored in MongoDB; a local token file is also written on the server)
- 💬 **Persistent chat history** — conversations are stored per-user in Redis and reloaded automatically
- 🤖 **Local inference** — powered by [vLLM](https://github.com/vllm-project/vllm) serving an OpenAI-compatible API on your own machine/GPU
- 🐳 **One-command infra** — Redis, MongoDB, and vLLM all run via Docker Compose
- 💻 **Terminal UX** — plain-text login, then interactive chat:

```text
[mahdi]: hello
chatbot: Hi! How can I help?
[mahdi]:
```

## Architecture

```text
┌──────────────┐         ┌───────────────┐
│  TCP client  │────────▶│  OldAI server │  (Go)
│  (nc/telnet) │         │  :ServerPort  │
└──────────────┘         └───────┬───────┘
                                 │
                 ┌───────────────┼───────────────┐
                 ▼               ▼               ▼
          ┌──────────┐   ┌──────────┐   ┌──────────┐
          │ MongoDB  │   │  Redis   │   │   vLLM   │
          │ accounts │   │  chat    │   │  model   │
          │ sessions │   │ history  │   │ inference│
          └──────────┘   └──────────┘   └──────────┘
```

The Go process is a TCP server. MongoDB, Redis, and vLLM run as Docker Compose services.

## Prerequisites

- [Go](https://go.dev/) 1.21+
- [Docker](https://www.docker.com/) and Docker Compose
- An NVIDIA GPU with drivers + [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) (required for the vLLM container)
- `nc` / `ncat` / `telnet` (or any TCP client) to connect

## Setup

**1. Clone the repo and create a `.env` file** in the project root:

```env
### Server ###
ServerHost=0.0.0.0
ServerPort=8080

### Redis ###
RedisHost=localhost
RedisPort=6479
RedisPassword=change-me
RedisSession=DefaultSession

### MongoDB ###
MongoHost=localhost
MongoPort=27017
MongoUser=admin
MongoPassword=change-me
MONGO_DB_NAME=OldAI
MONGO_URI=mongodb://admin:change-me@localhost:27017/?authSource=admin

### vLLM ###
VLLM_PORT=8003
VLLM_BASE_URL=http://localhost:8003
VLLM_TIMEOUT=60s
VLLM_MODEL=Qwen/Qwen2.5-0.5B-Instruct
VLLM_GPU_MEMORY_UTILIZATION=0.8
VLLM_MAX_MODEL_LEN=1024
```

> Use strong passwords for Redis and MongoDB. Never commit `.env`.

**2. Start the infrastructure:**

```bash
docker compose up -d
```

On first run, vLLM downloads and loads the model (can take a few minutes). Check status:

```bash
docker compose ps
```

Wait until `OldAIVLLM` is `healthy` and Mongo/Redis are `Up`.

**3. Run the server:**

```bash
go run cmd/server/main.go
# or, after building:
# go build -o oldai ./cmd/server && ./oldai
```

You should see:

```text
Starting the application...
Configuration loaded
Storage initialized: ...
Connected to MongoDB
listening on [::]:8080
```

## Usage

Connect with netcat (or any TCP client):

```bash
nc localhost 8080
```

### Login / register

```text
1. Login
2. Register
Choose an option (1 or 2): 2
Enter username: mahdi
Enter password: ********
Logged in as mahdi
[mahdi]:
```

### Chat

Type a message and press Enter. The assistant replies on the next line:

```text
[mahdi]: hey, what can you help me with?
chatbot: I'm OldAI — ask me anything.
[mahdi]: what is 2+2?
chatbot: 4
[mahdi]: exit
bye
```

- Empty lines are ignored.
- Type `exit` or `quit` to disconnect.
- History is stored per username in Redis and reloaded on the next session.

## Configuration notes

| Variable | Purpose |
|----------|---------|
| `ServerPort` | TCP port the Go server listens on |
| `MONGO_URI` | Must include user/password when Mongo is started with root credentials; prefer `?authSource=admin` |
| `RedisHost` / `RedisPort` / `RedisPassword` | Must match the Redis container and `--requirepass` |
| `VLLM_BASE_URL` | OpenAI-compatible base URL (no trailing `/v1/...`; the client appends the path) |
| `VLLM_MODEL` | Model id passed to vLLM / chat completions |

## Troubleshooting

| Symptom | What to check |
|---------|----------------|
| `connection refused` on `:27017` | `docker compose up -d mongo`, confirm `MongoPort` and `MONGO_URI` |
| `connection refused` on Redis | Redis container up, password matches `RedisPassword` |
| Blank terminal after connect | Session may have been restored silently; type a message or remove `~/.oldai/session.json` on the **server** host and reconnect |
| Chat errors / timeouts | Wait for vLLM healthcheck; confirm `VLLM_BASE_URL` and GPU memory settings |
| Login works but model fails | `docker compose logs vllm` |

## Project layout (typical)

```text
cmd/server/main.go          # TCP server entrypoint
internal/auth/              # users, sessions, bcrypt, local token helpers
internal/config/            # env loading
internal/storage/           # Redis chat history
internal/vllmclient/        # OpenAI-compatible chat client
internal/proto/             # optional JSON line protocol (not required for terminal chat)
docker-compose.yml          # redis, mongo, vllm
.env                        # local secrets (not committed)
```

## License

Add your license here.
```