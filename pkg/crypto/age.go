package crypto

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultKeyPath = "~/.config/sops/age/keys.txt"
)

// AgeKeypair represents an age encryption keypair
type AgeKeypair struct {
	PublicKey  string
	PrivateKey string
}

// GenerateAgeKeypair shells out to age-keygen to generate a keypair
func GenerateAgeKeypair() (*AgeKeypair, error) {
	// Check if age-keygen is installed
	if _, err := exec.LookPath("age-keygen"); err != nil {
		return nil, fmt.Errorf("age-keygen not found. Install from https://github.com/FiloSottile/age")
	}

	cmd := exec.Command("age-keygen")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("age-keygen failed: %w\nOutput: %s", err, string(output))
	}

	return parseAgeKeygenOutput(string(output))
}

func parseAgeKeygenOutput(output string) (*AgeKeypair, error) {
	var publicKey, privateKey string

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# Public key: ") {
			publicKey = strings.TrimPrefix(line, "# Public key: ")
			publicKey = strings.TrimSpace(publicKey)
		} else if strings.HasPrefix(line, "AGE-SECRET-KEY-") {
			privateKey = strings.TrimSpace(line)
		}
	}

	if publicKey == "" || privateKey == "" {
		return nil, fmt.Errorf("failed to parse age-keygen output")
	}

	return &AgeKeypair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// SavePrivateKey saves the private age key to disk
func SavePrivateKey(keypair *AgeKeypair, path string) error {
	// Expand ~ to home directory
	path = expandPath(path)

	// Create directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file already exists
	var existingContent []byte
	if _, err := os.Stat(path); err == nil {
		existingContent, _ = os.ReadFile(path)
	}

	// Write key file
	content := fmt.Sprintf("# created: %s\n# public key: %s\n%s\n",
		time.Now().Format(time.RFC3339),
		keypair.PublicKey,
		keypair.PrivateKey)

	// Append if file exists, otherwise create new
	if len(existingContent) > 0 {
		content = string(existingContent) + "\n" + content
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// GetPublicKeyFromFile extracts the public key from the private key file
func GetPublicKeyFromFile(path string) (string, error) {
	// Expand ~
	path = expandPath(path)

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("age key file not found at %s", path)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# public key: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# public key: ")), nil
		}
	}

	return "", fmt.Errorf("no public key found in %s", path)
}

// CheckAgeInstalled verifies age-keygen binary is available
func CheckAgeInstalled() error {
	if _, err := exec.LookPath("age-keygen"); err != nil {
		return fmt.Errorf("age-keygen not found. Install from https://github.com/FiloSottile/age")
	}
	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}
