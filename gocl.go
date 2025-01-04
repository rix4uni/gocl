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
const version = "v0.0.3"

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

func cloneAndInstall(repoURL, customPath string) error {
	// Step 1: Add https:// to the URL if missing
	if !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}

	// Step 2: Validate the repository URL
	if err := validateRepoURL(repoURL); err != nil {
		return fmt.Errorf("repository %q is invalid or inaccessible: %v", repoURL, err)
	}

	// Step 3: Extract the last part of the URL to determine the tool name
	toolName := filepath.Base(repoURL)

	// Step 4: Get the current directory to return to it later
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting original directory: %v", err)
	}

	// Step 5: Clone the repository (silent)
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL)
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil // Suppress error
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}

	// Step 6: Determine the directory to use
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

	// Step 7: Check if go.sum exists, and run go mod tidy if not
	if _, err := os.Stat("go.sum"); os.IsNotExist(err) {
	    cmd = exec.Command("go", "mod", "tidy")
	    cmd.Stdout = os.Stdout
	    cmd.Stderr = os.Stderr
	    if err := cmd.Run(); err != nil {
	        return fmt.Errorf("error running go mod tidy: %v", err)

	        // Step 8: Check if go.mod exists, and run go mod init if not
			if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
			    modulePath := strings.TrimPrefix(repoURL, "https://")
			    cmd = exec.Command("go", "mod", "init", modulePath)
			    cmd.Stdout = os.Stdout
			    cmd.Stderr = os.Stderr
			    if err := cmd.Run(); err != nil {
			        return fmt.Errorf("error running go mod init: %v", err)
			    }
			}
	    }
	}

	// Step 9: Run go install and display output
	cmd = exec.Command("go", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running go install: %v", err)
	}

	// Step 10: Change back to the original directory
	if err := os.Chdir(originalDir); err != nil {
		return fmt.Errorf("error returning to original directory: %v", err)
	}

	// Step 11: Remove the cloned repository (silent)
	if err := os.RemoveAll(toolName); err != nil {
		return fmt.Errorf("error removing cloned repository: %v", err)
	}

	return nil
}

func main() {
	// Define the flags
	inputFlag := pflag.StringP("input", "i", "", "URL or file containing URLs of the repository to install")
	customPathFlag := pflag.StringP("custom-path", "c", "", "Custom path to use for installation (e.g., cmd/interactsh-client).")
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
				// Call cloneAndInstall for each URL with the custom path
				if err := cloneAndInstall(repoURL, *customPathFlag); err != nil {
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
		if err := cloneAndInstall(repoURL, *customPathFlag); err != nil {
			fmt.Println("Error:", err)
		}
	}
}
