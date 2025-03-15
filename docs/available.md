# Available Command Documentation

The `available` command (with alias `avail`) allows you to browse and search the entire Ollama model library without downloading anything. This is useful for discovering models that you might want to pull to your Ollama server.

## Basic Usage

```bash
ollama-cli available
```

This will list all models available on ollama.com.

## Command Options

| Flag | Shorthand | Description |
|------|-----------|-------------|
| `--output` | `-o` | Output format (table, wide, json) |
| `--details` | `-d` | Show detailed information about models |
| `--filter` | `-f` | Filter models by name |
| `--timeout` | `-t` | Timeout in seconds for the HTTP request (default: 30) |

## Examples

### List All Available Models

```bash
# List all available models from Ollama's library
ollama-cli available
```

Example output:
```
NAME                SIZE                UPDATED
llama3              8b                  2 weeks ago
phi3                14b                 1 month ago
mistral             7b                  2 months ago
...
```

### Filter Models by Name

```bash
# Filter models by name
ollama-cli avail -f llama
```

Example output:
```
NAME                SIZE                UPDATED
llama3              8b                  2 weeks ago
llama2              7b,13b,70b          3 months ago
codellama           7b,13b,34b,70b      4 months ago
...
```

### Show Detailed Model Information

```bash
# Show detailed model information
ollama-cli avail --details
```

Example output:
```
NAME                SIZE                UPDATED             DESCRIPTION
llama3              8b                  2 weeks ago         Meta's latest open LLM with improved reasoning
phi3                14b                 1 month ago         Microsoft's state-of-the-art small language model
mistral             7b                  2 months ago        Mistral AI's efficient and powerful base model
...
```

### Wide Format Output

```bash
# Show all available information in a wide table
ollama-cli avail --output wide
```

Example output:
```
NAME                SIZE                PULLS               TAGS                UPDATED             DESCRIPTION
llama3              8b                  1.2M                latest,v1           2 weeks ago         Meta's latest open LLM with improved reasoning
phi3                14b                 800K                latest,v1           1 month ago         Microsoft's state-of-the-art small language model
mistral             7b                  1.5M                latest,v0.1,v0.2    2 months ago        Mistral AI's efficient and powerful base model
...
```

### JSON Output for Scripting

```bash
# Get output in JSON format for scripting
ollama-cli avail --output json
```

Example output:
```json
[
  {
    "name": "llama3",
    "size": "8b",
    "pulls": "1.2M",
    "tags": "latest,v1",
    "updated": "2 weeks ago",
    "description": "Meta's latest open LLM with improved reasoning"
  },
  {
    "name": "phi3",
    "size": "14b",
    "pulls": "800K",
    "tags": "latest,v1",
    "updated": "1 month ago",
    "description": "Microsoft's state-of-the-art small language model"
  },
  ...
]
```

### Adjust Timeout for Slow Connections

```bash
# Increase timeout for slow connections
ollama-cli avail --timeout 60
```

## Using with Other Commands

The `available` command works well in combination with other Ollama CLI commands:

### Find and Pull a Model

```bash
# Find models matching a specific name
ollama-cli avail -f phi

# Pull a specific model to your server
ollama-cli pull phi3:14b
```

### Check Available Models Before Pulling

```bash
# Check what versions of a model are available
ollama-cli avail -f llama3

# Pull the specific version you want
ollama-cli pull llama3:8b
```

## Troubleshooting

### Common Issues

1. **"Failed to fetch available models"**: Check your internet connection and try increasing the timeout value with `--timeout`.

2. **No models found**: If filtering returns no results, try a more general filter term or check the spelling.

3. **Slow response**: The command fetches data from ollama.com, which might be slow depending on your connection. Use the `--timeout` flag to increase the wait time.

For additional help, please refer to the project's GitHub repository or open an issue if you encounter persistent problems. 