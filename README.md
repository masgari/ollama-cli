# Ollama CLI

A command-line interface for interacting with a remote Ollama server.

## Features

- List models available on a remote Ollama server
- Configure connection to a remote Ollama server
- Remove models from the Ollama server
- More features coming soon!

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/masgari/ollama-cli.git
cd ollama-cli

# Build the binary
go build -o build/ollama-cli

# Install the binary (optional)
sudo cp build/ollama-cli /usr/local/bin/
```

### Build Options

```bash
# Build for your current platform
make build

# Build for all supported platforms
make build-all

# Build for a specific platform
GOOS=linux GOARCH=amd64 make build
```

The compiled binaries will be placed in the `build/` directory.

## Configuration

The CLI tool stores its configuration in `~/.ollama-cli/config.yaml`. The configuration file is created automatically on first run with default values.

### View Current Configuration

```bash
ollama-cli config
```

### Update Configuration

```bash
# Set the host and port
ollama-cli config --host remote-server.example.com --port 11434

# Or use the set command
ollama-cli config set host remote-server.example.com
ollama-cli config set port 11434
```

### Get Configuration Values

```bash
ollama-cli config get host
ollama-cli config get port
ollama-cli config get url
```

You can also override the configuration using command-line flags:

```bash
ollama-cli --host remote-server.example.com --port 11434 list
```

## Usage

### List Models

List all models available on the Ollama server:

```bash
ollama-cli list
# or
ollama-cli ls
```

#### Output Formats

```bash
# Default table format
ollama-cli list

# Show detailed information
ollama-cli list --details

# Wide format with all details
ollama-cli list --output wide

# JSON format
ollama-cli list --output json
```

### Remove Models

Remove a model from the Ollama server:

```bash
ollama-cli rm model-name
# or
ollama-cli delete model-name
# or
ollama-cli remove model-name
```

By default, the command will ask for confirmation before deleting the model. You can use the `--force` flag to skip the confirmation:

```bash
ollama-cli rm model-name --force
# or
ollama-cli rm model-name -f
```

### Version

Display the version of the CLI tool:

```bash
ollama-cli version
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 