package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

// prints the version message
const version = "v0.0.2"

func PrintVersion() {
	fmt.Printf("Current gocl version %s\n", version)
}

func cloneAndInstall(repoURL string) error {
	// Step 1: Add https:// to the URL if missing
	if !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}

	// Step 2: Extract the last part of the URL to determine the tool name
	toolName := filepath.Base(repoURL)

	// Step 3: Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gocl_*")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up the temporary directory afterward

	// Step 4: Clone the repository into the temporary directory
	clonePath := filepath.Join(tempDir, toolName)
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, clonePath)
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil // Suppress error
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}

	// Step 5: Determine the directory containing the main package
	// Check if a `cmd/<toolName>` directory exists
	cmdDir := filepath.Join(clonePath, "cmd", toolName)
	if _, err := os.Stat(cmdDir); err == nil {
		// Use the `cmd/<toolName>` directory
		if err := os.Chdir(cmdDir); err != nil {
			return fmt.Errorf("error changing directory to cmd/<toolName>: %v", err)
		}
	} else if err := os.Chdir(clonePath); err != nil {
		// Fallback to the root directory if `cmd/<toolName>` doesn't exist
		return fmt.Errorf("error changing directory: %v", err)
	}

	// Step 6: Run go install and display output
	cmd = exec.Command("go", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running go install: %v", err)
	}

	return nil
}

func main() {
	// Define the flags
	inputFlag := pflag.StringP("input", "i", "", "URL or file containing URLs of the repository to install")
	versionFlag := pflag.Bool("version", false, "Print the version of the tool and exit.")
	pflag.Parse()

	if *versionFlag {
		PrintVersion()
		return
	}

	// Check if the input flag is set
	if *inputFlag == "" {
		fmt.Println("Usage:")
		fmt.Println(" gocl -i github.com/rix4uni/gocl")
		fmt.Println(" gocl -i urls.txt")
		fmt.Println("\nurls.txt:")
		fmt.Println(" github.com/rix4uni/gocl")
		fmt.Println(" github.com/rix4uni/unew")
		return
	}

	// Check if the argument is a file or a URL
	if _, err := os.Stat(*inputFlag); err == nil {
		// If it's a file, read URLs from the file
		file, err := os.Open(*inputFlag)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			repoURL := strings.TrimSpace(scanner.Text())
			if repoURL != "" {
				if err := cloneAndInstall(repoURL); err != nil {
					fmt.Println("Error:", err)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
		}
	} else {
		// If it's a URL, process it directly
		repoURL := *inputFlag
		if err := cloneAndInstall(repoURL); err != nil {
			fmt.Println("Error:", err)
		}
	}
}
