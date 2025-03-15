# Ollama CLI

A lightweight command-line interface for managing remote Ollama servers without installing Ollama locally.

## Why Ollama CLI?

While Ollama provides its own CLI, it requires a local Ollama installation. Ollama CLI lets you manage remote Ollama servers from any machine without installing Ollama itself. Suitable for:

- Managing Ollama on headless servers
- Working with multiple Ollama instances across your network
- Accessing Ollama from environments where installing it isn't practical

## Quick Install
```bash
# macOS (Apple Silicon)
curl -L https://github.com/masgari/ollama-cli/releases/latest/download/ollama-cli-darwin-arm64 -o ollama-cli
# macOS (Intel)
curl -L https://github.com/masgari/ollama-cli/releases/latest/download/ollama-cli-darwin-amd64 -o ollama-cli
# Linux (x86_64)
curl -L https://github.com/masgari/ollama-cli/releases/latest/download/ollama-cli-linux-amd64 -o ollama-cli

# make executable
chmod +x ollama-cli
sudo mv ollama-cli /usr/local/bin/


# for Windows (x86_64)
curl -L https://github.com/masgari/ollama-cli/releases/latest/download/ollama-cli-windows-amd64.exe -o ollama-cli.exe
```

Or see [Installation from Source](#installation-from-source) below.

## Key Features

```bash
ollama-cli -h

ollama-cli is a command-line interface for interacting with a remote Ollama server.
It allows you to manage models, run inferences, and more.

Usage:
  ollama-cli [command]

Available Commands:
  available   List models available on ollama.com
  completion  Generate the autocompletion script for the specified shell
  config      Configure the Ollama CLI
  help        Help about any command
  list        List models available on the Ollama server
  pull        Pull a model from the Ollama server
  rm          Remove a model from the Ollama server
  chat        Chat with an Ollama model
  version     Display the version of the CLI tool

Flags:
      --config string        config file (default is $HOME/.ollama-cli/config.yaml)
  -c, --config-name string   config name to use (e.g. 'pc' for $HOME/.ollama-cli/pc.yaml)
  -h, --help                 help for ollama-cli
  -H, --host string          Ollama server host (default is localhost)
      --no-color             Disable color output
  -p, --port int             Ollama server port (default is 11434)
  -v, --verbose              verbose output

Use "ollama-cli [command] --help" for more information about a command.
```

### Discover Available Models

Browse and search the entire Ollama model library without downloading anything:

```bash
# List all available models from Ollama's library
ollama-cli available

# Filter models by name
ollama-cli avail -f deep
NAME                  SIZE                                   UPDATED
deepscaler           1.5b                                   3 weeks ago
deepseek-r1          1.5b,7b,8b,14b,32b,70b,671b          4 weeks ago
deepseek-v3          671b                                   1 month ago
deepseek-v2.5        236b                                   5 months ago
...

# Show detailed model information
ollama-cli avail --details
```

### Manage Multiple Ollama Servers

Create and switch between configurations for different Ollama servers:

```bash
# Set up configurations for different servers
ollama-cli config --host 192.168.1.90 # uses default config in ~/.ollama-cli/config.yaml

ollama-cli -c pi5 config --host 192.168.1.100 # uses config in ~/.ollama-cli/pi5.yaml
ollama-cli -c pc config --host 192.168.1.101 # uses config in ~/.ollama-cli/pc.yaml

# Use a specific configuration
ollama-cli -c pc list
```

### Manage Remote Models

Work with models on your remote Ollama server:

```bash

# List models on the server
ollama-cli list
NAME                   SIZE                MODIFIED
phi4-mini:3.8b       2.3 GB     2 days ago
llava-phi3:latest    2.7 GB     1 months ago
...

# check available models
ollama-cli available --filter smallthinker
NAME             SIZE          UPDATED
smallthinker   3b   1 month ago

# Pull a model to the server
ollama-cli pull smallthinker:3b
Pulling model 'smallthinker:3b'...
smallthinker:3b: [57.1%] [1967.8/3448.6 MB] pulling ad361f123f77

# Remove a model
ollama-cli rm smallthinker:3b
```

### Chat with Models

Interact with your LLM models:

```bash
# Simple chat with a model
ollama-cli chat llama3.2

# Chat with a specific input
ollama-cli chat llama3.2 --prompt "What is the capital of France?"
ollama-cli chat llama3.2 -p "What is the capital of France?"

# Chat with a model using an image
ollama-cli chat llama3.2 --prompt "What's in this image?" --image /path/to/image.jpg
ollama-cli chat llama3.2 -p "What's in this image?" -i /path/to/image.jpg

# Interactive chat mode
ollama-cli chat llama3.2 --interactive
ollama-cli chat llama3.2 -I

# Set custom parameters
ollama-cli chat llama3.2 --temperature 0.7 --system "You are a helpful assistant"
ollama-cli chat llama3.2 -t 0.7 -s "You are a helpful assistant"

# Save chat history to a file
ollama-cli chat llama3.2 --output-file chat_history.json

# Load previous chat history
ollama-cli chat llama3.2 --input-file chat_history.json

# Display statistics about token usage and generation time
ollama-cli chat llama3.2 --prompt "Hello" --stats --no-stream
```

In interactive mode, you can use special commands:
- Type `exit` to quit the chat session
- Type `save` to save the conversation
- Type `clear` to clear the chat history
- Type `temp 0.8` to change the temperature parameter
- Type `image /path/to/image.jpg` to send an image

> **Note**: The chat command is disabled by default for security reasons. When you first run it, you will be prompted to enable it.
> For detailed usage instructions and security considerations, see [Chat Documentation](docs/chat.md) and [Security Guidelines](docs/security.md).

### Flexible Output Formats

All commands support multiple output formats:

```bash
# Default table format
ollama-cli list

# Detailed view with all fields
ollama-cli list --output wide

# JSON output for scripting
ollama-cli list --output json
```

## Configuration

The CLI stores its configuration in `~/.ollama-cli/config.yaml`, created automatically on first run.

```bash
# View current configuration
ollama-cli config

# Update configuration
ollama-cli config --host remote-server.example.com --port 11434

# Override config for a single command
ollama-cli --host 192.168.1.200 list
```

## Installation from Source

```bash
# Clone the repository
git clone https://github.com/masgari/ollama-cli.git
cd ollama-cli

# Build for your platform
make build

# Or build for a specific platform
GOOS=linux GOARCH=amd64 make build
```

## Disclaimer

Ollama CLI is an open source project provided "as is" without warranty of any kind, express or implied. The authors and contributors of this project disclaim all liability for any damages or losses, including but not limited to direct, indirect, incidental, special, consequential, or similar damages arising from the use of this tool.

Users are responsible for ensuring their use of Ollama CLI complies with all applicable laws, regulations, and third-party terms of service. The authors are not responsible for any misuse of the tool or for any violations of Ollama's terms of service that may result from using this CLI.

By using Ollama CLI, you acknowledge and agree that you assume all risks associated with its use.

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.