## gocl

gocl is a tool similar to `go install` that installs Go tools from GitHub repositories with additional features for custom paths and binary management.

## Features

- ✅ Install Go tools from GitHub repositories
- ✅ Support for version specifiers (e.g., `@latest` - automatically stripped)
- ✅ Batch installation from file
- ✅ Custom path support for repositories with non-standard structures
- ✅ Save binaries to specific directories using `--location` flag
- ✅ Custom binary naming with `--output` flag
- ✅ Automatic directory detection (`v2/cmd/<tool>`, `cmd/<tool>`, or root)
- ✅ Automatic `go mod` initialization and tidying

## Installation

```
go install github.com/rix4uni/gocl@latest
```

## Download prebuilt binaries

```
wget https://github.com/rix4uni/gocl/releases/download/v0.0.5/gocl-linux-amd64-0.0.5.tgz
tar -xvzf gocl-linux-amd64-0.0.5.tgz
rm -rf gocl-linux-amd64-0.0.5.tgz
mv gocl ~/go/bin/gocl
```
Or download [binary release](https://github.com/rix4uni/gocl/releases) for your platform.

## Compile from source

```
git clone --depth 1 github.com/rix4uni/gocl.git
cd gocl; go install
```

## Usage

```
Usage of gocl:
  -c, --custom-path string   Custom path to use for installation (e.g., cmd/interactsh-client).
  -i, --input string         URL or file containing URLs of the repository to install
  -l, --location string      Directory to save the binary file (uses go build instead of go install).
  -o, --output string        Custom name for the output binary file (only used with --location).
      --version              Print the version of the tool and exit.
```

## Examples

### Basic Installation

```
# Install a tool (uses go install by default - saves to $GOPATH/bin)
gocl -i github.com/rix4uni/gocl

# Install with version specifier (@latest is automatically stripped)
gocl -i github.com/rix4uni/unew@latest

# Install from a file containing multiple URLs
gocl -i urls.txt
```

### Custom Path

```
# Install from a custom path within the repository
gocl -i github.com/projectdiscovery/interactsh -c cmd/interactsh-client
```

### Save Binary to Specific Directory

The `--location` flag creates a directory (if it doesn't exist) and saves the binary inside it. The binary is automatically named after the tool (last part of the repository URL).

```
# Create directory "bin" and save "gocl" binary inside it
gocl -i github.com/rix4uni/gocl -l ./bin
# Result: ./bin/gocl

# Create directory "garudrecon_binary" and save "ipfinder" binary inside it
gocl -i github.com/rix4uni/ipfinder -l garudrecon_binary
# Result: garudrecon_binary/ipfinder

# Use absolute path
gocl -i github.com/rix4uni/gocl -l /path/to/binaries
# Result: /path/to/binaries/gocl
```

### Custom Binary Name

Use the `--output` flag to specify a custom name for the binary file:

```
# Save binary with custom name
gocl -i github.com/rix4uni/ipfinder -l ./bin -o my-custom-tool
# Result: ./bin/my-custom-tool

# Combine custom path with location and output name
gocl -i github.com/projectdiscovery/interactsh -c cmd/interactsh-client -l ./tools -o interactsh
# Result: ./tools/interactsh
```

### Batch Installation from File

```
# urls.txt:
github.com/rix4uni/gocl
github.com/rix4uni/unew@latest
github.com/rix4uni/Gxss
github.com/projectdiscovery/chaos-client

# Install all tools from file (uses go install for each)
gocl -i urls.txt

# Install all tools to a specific directory
gocl -i urls.txt -l ./binaries
```

**Note:** When using `--location` flag, the tool uses `go build` instead of `go install`, giving you full control over where the binary is saved. The directory specified by `--location` is created automatically if it doesn't exist.

## How It Works

Instead of manually running these commands:
```
git clone --depth 1 https://github.com/rix4uni/gocl.git
cd gocl
go install
```

You can simply run:
```
gocl -i github.com/rix4uni/gocl
```

gocl automatically:
1. Clones the repository
2. Detects the correct directory structure
3. Handles `go mod` initialization if needed
4. Builds and installs the tool
5. Cleans up the cloned repository
