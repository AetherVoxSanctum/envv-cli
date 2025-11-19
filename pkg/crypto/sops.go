package crypto

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// EncryptWithSOPS shells out to sops binary for encryption
func EncryptWithSOPS(inputPath, outputPath string) error {
	// Check if sops is installed
	if _, err := exec.LookPath("sops"); err != nil {
		return fmt.Errorf("sops not found. Install from https://github.com/getsops/sops")
	}

	cmd := exec.Command("sops", "-e", inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sops encryption failed: %w\nOutput: %s", err, string(output))
	}

	if err := os.WriteFile(outputPath, output, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return nil
}

// DecryptWithSOPS shells out to sops binary for decryption
func DecryptWithSOPS(inputPath, outputPath string) error {
	// Check if sops is installed
	if _, err := exec.LookPath("sops"); err != nil {
		return fmt.Errorf("sops not found. Install from https://github.com/getsops/sops")
	}

	cmd := exec.Command("sops", "-d", inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a permission error
		if strings.Contains(string(output), "no valid decryption key") ||
			strings.Contains(string(output), "failed to decrypt") {
			return fmt.Errorf("you don't have permission to decrypt these secrets")
		}
		return fmt.Errorf("sops decryption failed: %w\nOutput: %s", err, string(output))
	}

	if err := os.WriteFile(outputPath, output, 0600); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}

// DecryptToMemory decrypts without writing to disk
func DecryptToMemory(inputPath string) ([]byte, error) {
	// Check if sops is installed
	if _, err := exec.LookPath("sops"); err != nil {
		return nil, fmt.Errorf("sops not found. Install from https://github.com/getsops/sops")
	}

	cmd := exec.Command("sops", "-d", inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "no valid decryption key") {
			return nil, fmt.Errorf("you don't have permission to decrypt these secrets")
		}
		return nil, fmt.Errorf("sops decryption failed: %w\nOutput: %s", err, string(output))
	}

	return output, nil
}

// GenerateSOPSConfig creates .sops.yaml with team public keys
func GenerateSOPSConfig(publicKeys []string) string {
	config := "creation_rules:\n"
	config += "  - path_regex: \\.env.*$\n"
	config += "    age: >-\n"

	for i, key := range publicKeys {
		if i > 0 {
			config += ","
		}
		config += "\n      " + key
	}
	config += "\n"

	return config
}

// ExtractSOPSMetadata parses encrypted JSON and extracts sops metadata
func ExtractSOPSMetadata(encryptedData []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(encryptedData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse encrypted data: %w", err)
	}

	sopsData, ok := data["sops"]
	if !ok {
		return nil, fmt.Errorf("no sops metadata found in encrypted file")
	}

	metadata, ok := sopsData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid sops metadata format")
	}

	return metadata, nil
}

// CheckSOPSInstalled verifies sops binary is available
func CheckSOPSInstalled() error {
	if _, err := exec.LookPath("sops"); err != nil {
		return fmt.Errorf("sops not found. Install from https://github.com/getsops/sops")
	}
	return nil
}
