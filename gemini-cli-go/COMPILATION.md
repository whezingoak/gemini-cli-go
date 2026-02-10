# Compilation and Usage Guide

## Prerequisites

- **Go**: Version 1.21 or later.
- **Git**: To clone the repository.
- **Google Cloud SDK**: For `gcloud auth application-default login` (optional but recommended).

## Building from Source

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/whezingoak/gemini-cli-go.git
    cd gemini-cli-go
    ```

2.  **Install dependencies**:
    ```bash
    go mod tidy
    ```

3.  **Build the binary**:
    ```bash
    go build -o gemini ./cmd/gemini
    ```

4.  **Run**:
    ```bash
    ./gemini chat
    ```

## Cross-Compilation

To build for different platforms:

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -o gemini-linux-amd64 ./cmd/gemini
```

### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o gemini.exe ./cmd/gemini
```

### macOS
```bash
GOOS=darwin GOARCH=arm64 go build -o gemini-macos-arm64 ./cmd/gemini
```

## Troubleshooting

### Authentication Errors

If you see "failed to find default credentials", ensure you have authenticated:

1.  Run `gcloud auth application-default login`.
2.  Or set `GEMINI_API_KEY` environment variable.

### Permission Errors

Ensure your Google Cloud Project has the Generative Language API enabled.
