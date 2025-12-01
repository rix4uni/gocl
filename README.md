## gocl

gocl is similar tool like `go install` command that install's go tools.

## Features

- ✅ Install Go tools from GitHub repositories
- ✅ Support for version specifiers (e.g., `@latest` - automatically stripped)
- ✅ Batch installation from file
- ✅ Custom path support for repositories with non-standard structures
- ✅ Save binaries to specific locations using `--location` flag
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
  -l, --location string      Specific location to save the binary file (uses go build instead of go install).
      --version              Print the version of the tool and exit.
```

## Examples

### Basic Installation
```
# Install a tool (uses go install by default)
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

### Save Binary to Specific Location
```
# Save binary to a specific file path
gocl -i github.com/rix4uni/gocl -l /path/to/my-binary

# Save binary to a directory (will be named after the tool)
gocl -i github.com/rix4uni/gocl -l /path/to/binaries/

# Combine custom path with location
gocl -i github.com/projectdiscovery/interactsh -c cmd/interactsh-client -l ./bin/
```

### Batch Installation from File
```
# urls.txt:
github.com/rix4uni/gocl
github.com/rix4uni/unew@latest
github.com/rix4uni/Gxss
github.com/projectdiscovery/chaos-client

# Note: When using a file, you can use --custom-path flag for all URLs
# gocl -i urls.txt -c cmd/chaos
```

**Note:** When using `--location` flag, the tool uses `go build` instead of `go install`, giving you full control over where the binary is saved.

#### You can do this manually but you need to run 3 commands
```
git clone --depth 1 https://github.com/rix4uni/gocl.git
cd gocl
go install
```