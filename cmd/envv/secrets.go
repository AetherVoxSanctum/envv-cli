package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/AetherVoxSanctum/envv-cli/v3/pkg/api"
	"github.com/AetherVoxSanctum/envv-cli/v3/pkg/config"
	"github.com/AetherVoxSanctum/envv-cli/v3/pkg/crypto"
)

// secretsCommand returns the secrets command with subcommands
func secretsCommand() cli.Command {
	return cli.Command{
		Name:    "secrets",
		Aliases: []string{"secret"},
		Usage:   "Secrets management commands",
		Subcommands: []cli.Command{
			{
				Name:      "push",
				Usage:     "Push encrypted secrets to the backend",
				ArgsUsage: "[file]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "env",
						Usage: "Environment (development, staging, production)",
					},
					cli.StringFlag{
						Name:  "format",
						Usage: "Format (dotenv, json, yaml)",
						Value: "dotenv",
					},
				},
				Action: secretsPushAction,
			},
			{
				Name:  "pull",
				Usage: "Pull and decrypt secrets from the backend",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "env",
						Usage: "Environment (development, staging, production)",
					},
					cli.StringFlag{
						Name:  "output",
						Usage: "Output file path",
					},
				},
				Action: secretsPullAction,
			},
			{
				Name:  "sync",
				Usage: "Sync local encrypted file with backend",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "env",
						Usage: "Environment (development, staging, production)",
					},
				},
				Action: secretsSyncAction,
			},
			{
				Name:  "list",
				Usage: "List secret versions for an environment",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "env",
						Usage: "Environment (development, staging, production)",
					},
				},
				Action: secretsListAction,
			},
			{
				Name:  "rotate",
				Usage: "Rotate encryption keys (re-encrypt for new team members)",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "env",
						Usage: "Environment (development, staging, production)",
					},
				},
				Action: secretsRotateAction,
			},
		},
	}
}

func secretsPushAction(c *cli.Context) error {
	// Get file path from args
	filePath := c.Args().First()
	if filePath == "" {
		return fmt.Errorf("file path is required")
	}

	env := c.String("env")
	format := c.String("format")

	// Load project config
	cfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	// Use default env if not specified
	if env == "" {
		env = cfg.DefaultEnv
	}

	// Validate environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvs[env] {
		return fmt.Errorf("env must be one of: development, staging, production")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Check if SOPS is installed
	if err := crypto.CheckSOPSInstalled(); err != nil {
		return err
	}

	fmt.Printf("🔐 Pushing secrets for environment: %s\n", env)

	// Get authenticated client
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Get project members to get their public keys
	membersResp, err := client.GetProjectMembers(ctx, cfg.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project members: %w", err)
	}

	// Extract public keys
	var publicKeys []string
	for _, member := range membersResp.Members {
		if member.AgePublicKey != "" {
			publicKeys = append(publicKeys, member.AgePublicKey)
		}
	}

	if len(publicKeys) == 0 {
		return fmt.Errorf("no members with age public keys found")
	}

	fmt.Printf("   Encrypting for %d team members...\n", len(publicKeys))

	// Generate .sops.yaml config
	sopsConfig := crypto.GenerateSOPSConfig(publicKeys)
	sopsConfigPath := ".sops.yaml"

	// Write .sops.yaml temporarily
	if err := os.WriteFile(sopsConfigPath, []byte(sopsConfig), 0644); err != nil {
		return fmt.Errorf("failed to write .sops.yaml: %w", err)
	}
	defer os.Remove(sopsConfigPath) // Clean up

	// Encrypt the file
	encryptedPath := filePath + ".encrypted"
	if err := crypto.EncryptWithSOPS(filePath, encryptedPath); err != nil {
		return fmt.Errorf("failed to encrypt file: %w", err)
	}
	defer os.Remove(encryptedPath) // Clean up

	// Read encrypted data
	encryptedData, err := os.ReadFile(encryptedPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Extract SOPS metadata
	metadata, err := crypto.ExtractSOPSMetadata(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to extract SOPS metadata: %w", err)
	}

	fmt.Println("   Uploading to backend...")

	// Push to backend
	req := api.PushSecretsRequest{
		EncryptedData: string(encryptedData),
		Format:        format,
		Environment:   env,
		SOPSMetadata:  metadata,
	}

	version, err := client.PushSecrets(ctx, cfg.ProjectID, req)
	if err != nil {
		return fmt.Errorf("failed to push secrets: %w", err)
	}

	fmt.Printf("\n✅ Secrets pushed successfully!\n")
	fmt.Printf("   Environment: %s\n", env)
	fmt.Printf("   Version:     %d\n", version.Version)
	fmt.Printf("   Size:        %d bytes\n", len(encryptedData))

	return nil
}

func secretsPullAction(c *cli.Context) error {
	env := c.String("env")
	outputPath := c.String("output")

	// Load project config
	cfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	// Use default env if not specified
	if env == "" {
		env = cfg.DefaultEnv
	}

	// Default output path
	if outputPath == "" {
		outputPath = fmt.Sprintf(".env.%s", env)
	}

	// Check if SOPS is installed
	if err := crypto.CheckSOPSInstalled(); err != nil {
		return err
	}

	fmt.Printf("🔓 Pulling secrets for environment: %s\n", env)

	// Get authenticated client
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Pull from backend
	secretsResp, err := client.PullSecrets(ctx, cfg.ProjectID, env)
	if err != nil {
		return fmt.Errorf("failed to pull secrets: %w", err)
	}

	fmt.Printf("   Version: %d\n", secretsResp.Version)
	fmt.Println("   Decrypting...")

	// Write encrypted data to temp file
	encryptedPath := outputPath + ".encrypted"
	if err := os.WriteFile(encryptedPath, []byte(secretsResp.EncryptedData), 0600); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}
	defer os.Remove(encryptedPath) // Clean up

	// Decrypt with SOPS
	if err := crypto.DecryptWithSOPS(encryptedPath, outputPath); err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	fmt.Printf("\n✅ Secrets pulled and decrypted!\n")
	fmt.Printf("   Output:  %s\n", outputPath)
	fmt.Printf("   Version: %d\n", secretsResp.Version)
	fmt.Printf("   Updated: %s\n", secretsResp.UpdatedAt)

	return nil
}

func secretsSyncAction(c *cli.Context) error {
	env := c.String("env")

	// Load project config
	cfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	// Use default env if not specified
	if env == "" {
		env = cfg.DefaultEnv
	}

	fmt.Printf("🔄 Syncing secrets for environment: %s\n", env)

	// Pull latest
	pullCtx := cli.NewContext(c.App, c.FlagSet, c)
	pullCtx.Set("env", env)
	pullCtx.Set("output", fmt.Sprintf(".env.%s", env))

	if err := secretsPullAction(pullCtx); err != nil {
		// If no secrets exist yet, that's okay
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
		fmt.Println("   No remote secrets found, will create new version")
	}

	// Push current version
	pushCtx := cli.NewContext(c.App, c.FlagSet, c)
	pushCtx.Set("env", env)

	filePath := fmt.Sprintf(".env.%s", env)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s. Create it first", filePath)
	}

	// Set args for push
	args := cli.Args([]string{filePath})
	pushCtx.App.Metadata = map[string]interface{}{"args": args}

	if err := secretsPushAction(pushCtx); err != nil {
		return err
	}

	fmt.Println("\n✅ Sync complete!")

	return nil
}

func secretsListAction(c *cli.Context) error {
	env := c.String("env")

	// Load project config
	cfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	// Use default env if not specified
	if env == "" {
		env = cfg.DefaultEnv
	}

	// Get authenticated client
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	versions, err := client.ListSecretVersions(ctx, cfg.ProjectID, env)
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Printf("No versions found for environment: %s\n", env)
		return nil
	}

	fmt.Printf("\n📜 Secret Versions for %s (%d):\n\n", env, len(versions))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VERSION\tSIZE\tCREATED\tCREATED BY")
	fmt.Fprintln(w, "-------\t----\t-------\t----------")

	for _, v := range versions {
		fmt.Fprintf(w, "%d\t%d bytes\t%s\t%s\n",
			v.Version, v.Size, v.CreatedAt, v.CreatedBy)
	}

	w.Flush()
	fmt.Println()

	return nil
}

func secretsRotateAction(c *cli.Context) error {
	env := c.String("env")

	// Load project config
	cfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	// Use default env if not specified
	if env == "" {
		env = cfg.DefaultEnv
	}

	fmt.Printf("🔄 Rotating encryption keys for environment: %s\n", env)

	// Get authenticated client
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Pull current secrets (decrypted)
	fmt.Println("   1. Pulling current secrets...")
	secretsResp, err := client.PullSecrets(ctx, cfg.ProjectID, env)
	if err != nil {
		return fmt.Errorf("failed to pull secrets: %w", err)
	}

	// Decrypt to memory
	encryptedPath := ".envv/temp.encrypted"
	decryptedPath := ".envv/temp.decrypted"

	// Ensure .envv directory exists
	os.MkdirAll(".envv", 0700)
	defer os.Remove(encryptedPath)
	defer os.Remove(decryptedPath)

	if err := os.WriteFile(encryptedPath, []byte(secretsResp.EncryptedData), 0600); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	if err := crypto.DecryptWithSOPS(encryptedPath, decryptedPath); err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	// Get updated member list
	fmt.Println("   2. Fetching current team members...")
	membersResp, err := client.GetProjectMembers(ctx, cfg.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project members: %w", err)
	}

	// Extract public keys
	var publicKeys []string
	for _, member := range membersResp.Members {
		if member.AgePublicKey != "" {
			publicKeys = append(publicKeys, member.AgePublicKey)
		}
	}

	if len(publicKeys) == 0 {
		return fmt.Errorf("no members with age public keys found")
	}

	fmt.Printf("   3. Re-encrypting for %d team members...\n", len(publicKeys))

	// Generate new .sops.yaml config
	sopsConfig := crypto.GenerateSOPSConfig(publicKeys)
	sopsConfigPath := ".sops.yaml"

	if err := os.WriteFile(sopsConfigPath, []byte(sopsConfig), 0644); err != nil {
		return fmt.Errorf("failed to write .sops.yaml: %w", err)
	}
	defer os.Remove(sopsConfigPath)

	// Re-encrypt with new keys
	newEncryptedPath := ".envv/temp.reencrypted"
	defer os.Remove(newEncryptedPath)

	if err := crypto.EncryptWithSOPS(decryptedPath, newEncryptedPath); err != nil {
		return fmt.Errorf("failed to re-encrypt: %w", err)
	}

	// Read re-encrypted data
	encryptedData, err := os.ReadFile(newEncryptedPath)
	if err != nil {
		return fmt.Errorf("failed to read re-encrypted file: %w", err)
	}

	// Extract new SOPS metadata
	metadata, err := crypto.ExtractSOPSMetadata(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to extract SOPS metadata: %w", err)
	}

	fmt.Println("   4. Pushing rotated secrets...")

	// Push back to backend
	req := api.PushSecretsRequest{
		EncryptedData: string(encryptedData),
		Format:        secretsResp.Format,
		Environment:   env,
		SOPSMetadata:  metadata,
	}

	version, err := client.PushSecrets(ctx, cfg.ProjectID, req)
	if err != nil {
		return fmt.Errorf("failed to push rotated secrets: %w", err)
	}

	fmt.Printf("\n✅ Secrets rotated successfully!\n")
	fmt.Printf("   Environment: %s\n", env)
	fmt.Printf("   New Version: %d\n", version.Version)
	fmt.Printf("   Encrypted for: %d members\n", len(publicKeys))

	return nil
}
