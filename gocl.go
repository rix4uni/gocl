package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// prints the version message
const version = "v0.0.5"

func PrintVersion() {
	fmt.Printf("Current gocl version %s\n", version)
}

// validateRepoURL checks if the repository URL is reachable.
func validateRepoURL(repoURL string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(repoURL)
	if err != nil {
		return fmt.Errorf("repository validation failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("repository validation failed: received status code %d", resp.StatusCode)
	}
	return nil
}

func cloneAndInstall(repoURL, customPath, location, outputName string) error {
	// Step 1: Remove version specifiers (e.g., @latest) from the URL
	if idx := strings.Index(repoURL, "@"); idx != -1 {
		repoURL = repoURL[:idx]
	}

	// Step 2: Add https:// to the URL if missing
	if !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}

	// Step 3: Validate the repository URL
	if err := validateRepoURL(repoURL); err != nil {
		return fmt.Errorf("repository %q is invalid or inaccessible: %v", repoURL, err)
	}

	// Step 4: Extract the last part of the URL to determine the tool name
	toolName := filepath.Base(repoURL)

	// Step 5: Get the current directory to return to it later
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting original directory: %v", err)
	}

	// Step 5.5: Resolve location path relative to original directory (before changing dirs)
	var resolvedLocation string
	if location != "" {
		// Determine the binary name (use outputName if provided, otherwise toolName)
		binaryName := toolName
		if outputName != "" {
			binaryName = outputName
		}

		// Resolve location path (always treat as directory)
		if filepath.IsAbs(location) {
			resolvedLocation = filepath.Join(location, binaryName)
		} else {
			resolvedLocation = filepath.Join(originalDir, location, binaryName)
		}
		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(resolvedLocation), 0755); err != nil {
			return fmt.Errorf("error creating output directory: %v", err)
		}
	}

	// Step 6: Clone the repository (silent)
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL)
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil // Suppress error
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}

	// Step 7: Determine the directory to use
	clonedDir := filepath.Join(originalDir, toolName)
	v2CmdDir := filepath.Join(clonedDir, "v2", "cmd", toolName)
	cmdDir := filepath.Join(clonedDir, "cmd", toolName)

	// Priority order: custom path > v2/cmd/<toolName> > cmd/<toolName> > root
	if customPath != "" {
		customPathDir := filepath.Join(clonedDir, customPath)
		if _, err := os.Stat(customPathDir); err == nil {
			if err := os.Chdir(customPathDir); err != nil {
				return fmt.Errorf("error changing directory to custom path %q: %v", customPath, err)
			}
		} else {
			return fmt.Errorf("custom path %q does not exist", customPath)
		}
	} else if _, err := os.Stat(v2CmdDir); err == nil {
		// Use the `v2/cmd/<toolName>` directory
		if err := os.Chdir(v2CmdDir); err != nil {
			return fmt.Errorf("error changing directory to v2/cmd/<toolName>: %v", err)
		}
	} else if _, err := os.Stat(cmdDir); err == nil {
		// Fallback to `cmd/<toolName>` directory if it exists
		if err := os.Chdir(cmdDir); err != nil {
			return fmt.Errorf("error changing directory to cmd/<toolName>: %v", err)
		}
	} else if err := os.Chdir(clonedDir); err != nil {
		// Fallback to the root directory if neither exists
		return fmt.Errorf("error changing directory to root: %v", err)
	}

	// Step 8: Check if go.sum exists, and run go mod tidy if not
	if _, err := os.Stat("go.sum"); os.IsNotExist(err) {
		cmd = exec.Command("go", "mod", "tidy")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("go mod tidy failed: %v\n", err)

			// Step 9: Check if go.mod exists, and run go mod init if not
			if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
				modulePath := strings.TrimPrefix(repoURL, "https://")
				cmd = exec.Command("go", "mod", "init", modulePath)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("error running go mod init: %v", err)
				}

				// Retry go mod tidy after initializing go.mod
				cmd = exec.Command("go", "mod", "tidy")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("error running go mod tidy after go mod init: %v", err)
				}
			} else {
				return fmt.Errorf("error: go.mod exists but go mod tidy failed: %v", err)
			}
		}
	}

	// Step 10: Build or install the binary
	if location != "" {
		// Use go build to save binary to specific location (use resolved absolute path)
		cmd = exec.Command("go", "build", "-o", resolvedLocation)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running go build: %v", err)
		}
		fmt.Printf("Binary saved to: %s\n", resolvedLocation)
	} else {
		// Use go install (default behavior)
		cmd = exec.Command("go", "install")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running go install: %v", err)
		}
	}

	// Step 11: Change back to the original directory
	if err := os.Chdir(originalDir); err != nil {
		return fmt.Errorf("error returning to original directory: %v", err)
	}

	// Step 12: Remove the cloned repository (silent)
	if err := os.RemoveAll(toolName); err != nil {
		return fmt.Errorf("error removing cloned repository: %v", err)
	}

	return nil
}

func main() {
	// Define the flags
	inputFlag := pflag.StringP("input", "i", "", "URL or file containing URLs of the repository to install")
	customPathFlag := pflag.StringP("custom-path", "c", "", "Custom path to use for installation (e.g., cmd/interactsh-client).")
	locationFlag := pflag.StringP("location", "l", "", "Directory to save the binary file (uses go build instead of go install).")
	outputFlag := pflag.StringP("output", "o", "", "Custom name for the output binary file (only used with --location).")
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
		fmt.Println(" gocl -i github.com/projectdiscovery/interactsh -c cmd/interactsh-client")
		fmt.Println(" gocl -i github.com/rix4uni/gocl -l ./bin")
		fmt.Println(" gocl -i github.com/rix4uni/ipfinder -l ./bin -o custom-name")
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
				// Call cloneAndInstall for each URL with the custom path, location, and output name
				if err := cloneAndInstall(repoURL, *customPathFlag, *locationFlag, *outputFlag); err != nil {
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
		if err := cloneAndInstall(repoURL, *customPathFlag, *locationFlag, *outputFlag); err != nil {
			fmt.Println("Error:", err)
		}
	}
}
