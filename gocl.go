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

func cloneAndInstall(repoURL string) error {
	// Step 1: Add https:// to the URL if missing
	if !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}

	// Step 2: Extract the last part of the URL to determine the tool name
	toolName := filepath.Base(repoURL)

	// Step 3: Get the current directory to return to it later
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting original directory: %v", err)
	}

	// Step 4: Clone the repository (silent)
	cmd := exec.Command("git", "clone", repoURL)
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil // Suppress error
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}

	// Step 5: Determine the directory containing the main package
	clonedDir := filepath.Join(originalDir, toolName)

	// Check if a `cmd/<toolName>` directory exists
	cmdDir := filepath.Join(clonedDir, "cmd", toolName)
	if _, err := os.Stat(cmdDir); err == nil {
		// Use the `cmd/<toolName>` directory
		if err := os.Chdir(cmdDir); err != nil {
			return fmt.Errorf("error changing directory to cmd/<toolName>: %v", err)
		}
	} else if err := os.Chdir(clonedDir); err != nil {
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

	// Step 7: Change back to the original directory
	if err := os.Chdir(originalDir); err != nil {
		return fmt.Errorf("error returning to original directory: %v", err)
	}

	// Step 8: Remove the cloned repository (silent)
	if err := os.RemoveAll(toolName); err != nil {
		return fmt.Errorf("error removing cloned repository: %v", err)
	}

	return nil
}

func main() {
	// Define the -i or --input flag
	inputFlag := pflag.StringP("input", "i", "", "URL or file containing URLs of the repository to install")

	// Parse the flags
	pflag.Parse()

	// Check if the input flag is set
	if *inputFlag == "" {
		fmt.Println("Usage: gocl -i repo-url/file-with-urls")
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
