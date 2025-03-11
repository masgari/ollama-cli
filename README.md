# Ollama CLI

A lightweight command-line interface for managing remote Ollama servers without installing Ollama locally.

## Why Ollama CLI?

While Ollama provides its own CLI, it requires a local Ollama installation. Ollama CLI lets you manage remote Ollama servers from any machine without installing Ollama itself. Perfect for:

- Managing Ollama on headless servers
- Working with multiple Ollama instances across your network
- Accessing Ollama from environments where installing it isn't practical

## Quick Install
```bash
# macOS (Apple Silicon)
curl -L https://github.com/masgari/ollama-cli/releases/download/v0.0.1/ollama-cli-darwin-arm64 -o ollama-cli
# macOS (Intel)
curl -L https://github.com/masgari/ollama-cli/releases/download/v0.0.1/ollama-cli-darwin-amd64 -o ollama-cli
# Linux (x86_64)
curl -L https://github.com/masgari/ollama-cli/releases/download/v0.0.1/ollama-cli-linux-amd64 -o ollama-cli

# make executable
chmod +x ollama-cli
sudo mv ollama-cli /usr/local/bin/


# for Windows (x86_64)
curl -L https://github.com/masgari/ollama-cli/releases/download/v0.0.1/ollama-cli-windows-amd64.exe -o ollama-cli.exe
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

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.