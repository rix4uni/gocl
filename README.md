## gocl

gocl is similar tool like `go install` command that install's go tools.

## Installation
```
go install github.com/rix4uni/gocl@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/gocl/releases/download/v0.0.2/gocl-linux-amd64-0.0.2.tgz
tar -xvzf gocl-linux-amd64-0.0.2.tgz
rm -rf gocl-linux-amd64-0.0.2.tgz
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
  -i, --input string   URL or file containing URLs of the repository to install
      --version        Print the version of the tool and exit.
```

## Examples
```
Usage:
 gocl -i github.com/rix4uni/gocl
 gocl -i urls.txt

Custom Path:
 gocl -i github.com/projectdiscovery/chaos-client -c "cmd/chaos"

urls.txt:
 github.com/rix4uni/gocl
 github.com/rix4uni/unew
 github.com/rix4uni/Gxss
```

#### You can do this manually but you need to run 3 commands
```
git clone --depth 1 https://github.com/rix4uni/gocl.git
cd gocl
go install
```