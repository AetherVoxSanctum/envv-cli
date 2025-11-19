package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/AetherVoxSanctum/envv-cli/v3/pkg/api"
	"github.com/AetherVoxSanctum/envv-cli/v3/pkg/config"
)

// projectCommand returns the project command with subcommands
func projectCommand() cli.Command {
	return cli.Command{
		Name:    "project",
		Aliases: []string{"proj"},
		Usage:   "Project management commands",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a new project",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "org-id",
						Usage: "Organization ID",
					},
					cli.StringFlag{
						Name:  "name",
						Usage: "Project name",
					},
					cli.StringFlag{
						Name:  "description",
						Usage: "Project description",
					},
				},
				Action: projectCreateAction,
			},
			{
				Name:  "list",
				Usage: "List projects in an organization",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "org-id",
						Usage: "Organization ID",
					},
				},
				Action: projectListAction,
			},
			{
				Name:  "init",
				Usage: "Initialize current directory as a project",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "org-id",
						Usage: "Organization ID",
					},
					cli.StringFlag{
						Name:  "project-id",
						Usage: "Project ID",
					},
					cli.StringFlag{
						Name:  "default-env",
						Usage: "Default environment (development, staging, production)",
						Value: "development",
					},
				},
				Action: projectInitAction,
			},
			{
				Name:   "status",
				Usage:  "Show current project configuration",
				Action: projectStatusAction,
			},
			{
				Name:  "members",
				Usage: "List project members",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "project-id",
						Usage: "Project ID (uses .envv/config.yaml if not specified)",
					},
				},
				Action: projectMembersAction,
			},
		},
	}
}

func projectCreateAction(c *cli.Context) error {
	orgID := c.String("org-id")
	name := c.String("name")
	description := c.String("description")

	// Prompt for missing values
	if orgID == "" {
		fmt.Print("Organization ID: ")
		fmt.Scanln(&orgID)
	}
	if name == "" {
		fmt.Print("Project name: ")
		fmt.Scanln(&name)
	}

	if orgID == "" || name == "" {
		return fmt.Errorf("org-id and name are required")
	}

	// Get authenticated client
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	fmt.Println("📦 Creating project...")

	req := api.CreateProjectRequest{
		Name:        name,
		Description: description,
	}

	project, err := client.CreateProject(ctx, orgID, req)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("\n✅ Project created successfully!\n")
	fmt.Printf("   ID:   %s\n", project.ID)
	fmt.Printf("   Name: %s\n", project.Name)
	fmt.Printf("   Slug: %s\n", project.Slug)
	fmt.Printf("\nTo initialize this directory, run:\n")
	fmt.Printf("   envv project init --org-id=%s --project-id=%s\n", orgID, project.ID)

	return nil
}

func projectListAction(c *cli.Context) error {
	orgID := c.String("org-id")

	if orgID == "" {
		return fmt.Errorf("--org-id is required")
	}

	// Get authenticated client
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	projects, err := client.ListProjects(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found. Create one with 'envv project create'")
		return nil
	}

	fmt.Printf("\n📋 Projects (%d):\n\n", len(projects))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSLUG\tMEMBERS\tCREATED")
	fmt.Fprintln(w, "--\t----\t----\t-------\t-------")

	for _, proj := range projects {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			proj.ID, proj.Name, proj.Slug, proj.MemberCount, proj.CreatedAt)
	}

	w.Flush()
	fmt.Println()

	return nil
}

func projectInitAction(c *cli.Context) error {
	orgID := c.String("org-id")
	projectID := c.String("project-id")
	defaultEnv := c.String("default-env")

	// Prompt for missing values
	if orgID == "" {
		fmt.Print("Organization ID: ")
		fmt.Scanln(&orgID)
	}
	if projectID == "" {
		fmt.Print("Project ID: ")
		fmt.Scanln(&projectID)
	}

	if orgID == "" || projectID == "" {
		return fmt.Errorf("org-id and project-id are required")
	}

	// Validate default environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvs[defaultEnv] {
		return fmt.Errorf("default-env must be one of: development, staging, production")
	}

	// Check if project config already exists
	if config.ProjectConfigExists() {
		return fmt.Errorf("project already initialized in this directory (.envv/config.yaml exists)")
	}

	// Get authenticated client to verify project exists
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Get project details
	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get organization details
	org, err := client.GetOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Create project config
	cfg := &config.ProjectConfig{
		OrganizationID:   orgID,
		OrganizationName: org.Name,
		ProjectID:        projectID,
		ProjectName:      project.Name,
		DefaultEnv:       defaultEnv,
	}

	if err := config.SaveProjectConfig(cfg); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	fmt.Printf("✅ Project initialized successfully!\n\n")
	fmt.Printf("   Organization: %s (%s)\n", org.Name, orgID)
	fmt.Printf("   Project:      %s (%s)\n", project.Name, projectID)
	fmt.Printf("   Default Env:  %s\n", defaultEnv)
	fmt.Printf("   Config:       %s\n", config.ProjectConfigFile)
	fmt.Printf("\nYou can now push/pull secrets with:\n")
	fmt.Printf("   envv secrets push .env.%s\n", defaultEnv)
	fmt.Printf("   envv secrets pull %s\n", defaultEnv)

	return nil
}

func projectStatusAction(c *cli.Context) error {
	// Load project config
	cfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	fmt.Println("Current Project:")
	fmt.Printf("  Organization ID:   %s\n", cfg.OrganizationID)
	fmt.Printf("  Organization Name: %s\n", cfg.OrganizationName)
	fmt.Printf("  Project ID:        %s\n", cfg.ProjectID)
	fmt.Printf("  Project Name:      %s\n", cfg.ProjectName)
	fmt.Printf("  Default Env:       %s\n", cfg.DefaultEnv)
	fmt.Printf("  Config File:       %s\n", config.ProjectConfigFile)

	return nil
}

func projectMembersAction(c *cli.Context) error {
	projectID := c.String("project-id")

	// If no project ID specified, load from config
	if projectID == "" {
		cfg, err := config.LoadProjectConfig()
		if err != nil {
			return err
		}
		projectID = cfg.ProjectID
	}

	// Get authenticated client
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	resp, err := client.GetProjectMembers(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project members: %w", err)
	}

	if len(resp.Members) == 0 {
		fmt.Println("No members found")
		return nil
	}

	fmt.Printf("\n👥 Project Members (%d):\n\n", len(resp.Members))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EMAIL\tNAME\tPERMISSION\tAGE PUBLIC KEY")
	fmt.Fprintln(w, "-----\t----\t----------\t--------------")

	for _, member := range resp.Members {
		// Truncate public key for display
		pubKey := member.AgePublicKey
		if len(pubKey) > 20 {
			pubKey = pubKey[:20] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			member.Email, member.Name, member.Permission, pubKey)
	}

	w.Flush()
	fmt.Println()

	return nil
}
