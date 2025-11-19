package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/AetherVoxSanctum/envv-cli/v3/pkg/api"
	"github.com/AetherVoxSanctum/envv-cli/v3/pkg/config"
	"github.com/AetherVoxSanctum/envv-cli/v3/pkg/crypto"
)

// authCommand returns the auth command with subcommands
func authCommand() cli.Command {
	return cli.Command{
		Name:  "auth",
		Usage: "Authentication commands",
		Subcommands: []cli.Command{
			{
				Name:  "register",
				Usage: "Register a new account with age key generation",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "email",
						Usage: "Email address",
					},
					cli.StringFlag{
						Name:  "password",
						Usage: "Password",
					},
					cli.StringFlag{
						Name:  "name",
						Usage: "Full name",
					},
					cli.StringFlag{
						Name:  "api-url",
						Usage: "API base URL",
						Value: "https://api.envv.sh",
					},
				},
				Action: registerAction,
			},
			{
				Name:  "login",
				Usage: "Login to existing account",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "email",
						Usage: "Email address",
					},
					cli.StringFlag{
						Name:  "password",
						Usage: "Password",
					},
					cli.StringFlag{
						Name:  "api-url",
						Usage: "API base URL",
						Value: "https://api.envv.sh",
					},
				},
				Action: loginAction,
			},
			{
				Name:   "logout",
				Usage:  "Logout and clear credentials",
				Action: logoutAction,
			},
			{
				Name:   "whoami",
				Usage:  "Show current user info",
				Action: whoamiAction,
			},
		},
	}
}

func registerAction(c *cli.Context) error {
	// Get flags
	email := c.String("email")
	password := c.String("password")
	name := c.String("name")
	apiURL := c.String("api-url")

	// Prompt for missing values
	if email == "" {
		fmt.Print("Email: ")
		fmt.Scanln(&email)
	}
	if password == "" {
		fmt.Print("Password: ")
		fmt.Scanln(&password)
	}
	if name == "" {
		fmt.Print("Full name: ")
		fmt.Scanln(&name)
	}

	// Validate inputs
	if email == "" || password == "" || name == "" {
		return fmt.Errorf("email, password, and name are required")
	}

	// Check if age-keygen is installed
	if err := crypto.CheckAgeInstalled(); err != nil {
		return err
	}

	fmt.Println("🔑 Generating age encryption keypair...")

	// Generate age keypair
	keypair, err := crypto.GenerateAgeKeypair()
	if err != nil {
		return fmt.Errorf("failed to generate age keypair: %w", err)
	}

	fmt.Printf("✓ Age public key: %s\n", keypair.PublicKey)

	// Save private key to default location
	if err := crypto.SavePrivateKey(keypair, crypto.DefaultKeyPath); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	fmt.Printf("✓ Private key saved to %s\n", crypto.DefaultKeyPath)

	// Register with backend
	client := api.NewClient(apiURL)
	ctx := context.Background()

	fmt.Println("\n🚀 Registering account...")

	req := api.RegisterEnhancedRequest{
		Email:        email,
		Password:     password,
		Name:         name,
		AgePublicKey: keypair.PublicKey,
	}

	authResp, err := client.RegisterEnhanced(ctx, req)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	// Save credentials
	creds := &config.Credentials{
		AccessToken: authResp.AccessToken,
		UserID:      authResp.User.ID,
		Email:       authResp.User.Email,
		ExpiresAt:   authResp.ExpiresAt,
	}

	if err := config.SaveCredentials(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Printf("\n✅ Successfully registered and logged in as %s\n", authResp.User.Email)
	fmt.Printf("   User ID: %s\n", authResp.User.ID)
	fmt.Printf("   Age Public Key: %s\n", authResp.User.AgePublicKey)

	return nil
}

func loginAction(c *cli.Context) error {
	// Get flags
	email := c.String("email")
	password := c.String("password")
	apiURL := c.String("api-url")

	// Prompt for missing values
	if email == "" {
		fmt.Print("Email: ")
		fmt.Scanln(&email)
	}
	if password == "" {
		fmt.Print("Password: ")
		fmt.Scanln(&password)
	}

	// Validate inputs
	if email == "" || password == "" {
		return fmt.Errorf("email and password are required")
	}

	// Login with backend
	client := api.NewClient(apiURL)
	ctx := context.Background()

	fmt.Println("🔐 Logging in...")

	req := api.LoginRequest{
		Email:    email,
		Password: password,
	}

	authResp, err := client.Login(ctx, req)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Save credentials
	creds := &config.Credentials{
		AccessToken: authResp.AccessToken,
		UserID:      authResp.User.ID,
		Email:       authResp.User.Email,
		ExpiresAt:   authResp.ExpiresAt,
	}

	if err := config.SaveCredentials(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Printf("\n✅ Successfully logged in as %s\n", authResp.User.Email)
	fmt.Printf("   User ID: %s\n", authResp.User.ID)

	return nil
}

func logoutAction(c *cli.Context) error {
	if err := config.ClearCredentials(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	fmt.Println("✅ Successfully logged out")
	return nil
}

func whoamiAction(c *cli.Context) error {
	// Load credentials
	creds, err := config.LoadCredentials()
	if err != nil {
		return err
	}

	// Get API URL from environment or use default
	apiURL := os.Getenv("ENVV_API_URL")
	if apiURL == "" {
		apiURL = "https://api.envv.sh"
	}

	// Get current user info
	client := api.NewClient(apiURL)
	client.SetToken(creds.AccessToken)
	ctx := context.Background()

	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	fmt.Println("Current User:")
	fmt.Printf("  Email:          %s\n", user.Email)
	fmt.Printf("  Name:           %s\n", user.Name)
	fmt.Printf("  User ID:        %s\n", user.ID)
	fmt.Printf("  Age Public Key: %s\n", user.AgePublicKey)
	fmt.Printf("  Created:        %s\n", user.CreatedAt)

	return nil
}
