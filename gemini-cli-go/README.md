# Gemini CLI (Go)

A faithful port of the official Google Gemini CLI to Go.

## Features

- **Interactive Chat**: Chat with Gemini 1.5 Pro directly from your terminal.
- **Markdown Rendering**: Beautifully formatted responses with syntax highlighting.
- **Authentication**: Supports Google Cloud Application Default Credentials (ADC) and API Keys.
- **Configuration**: Easy configuration via `settings.json`.

## Installation

```bash
go install github.com/whezingoak/gemini-cli-go/cmd/gemini@latest
```

## Usage

### Authentication

1.  **API Key**: Set the `GEMINI_API_KEY` environment variable.
    ```bash
    export GEMINI_API_KEY="your-api-key"
    ```
2.  **Application Default Credentials**:
    ```bash
    gcloud auth application-default login
    ```

### Chat

Start an interactive chat session:

```bash
gemini chat
```

### Configuration

Configuration is stored in `~/.config/gemini-cli/settings.json`.

Default settings:
```json
{
  "model": "gemini-1.5-pro",
  "maxOutputTokens": 8192,
  "temperature": 0.7,
  "location": "us-central1"
}
```

## Development

See [COMPILATION.md](COMPILATION.md) for build instructions.
