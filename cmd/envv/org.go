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

// orgCommand returns the organization command with subcommands
func orgCommand() cli.Command {
	return cli.Command{
		Name:    "org",
		Aliases: []string{"organization"},
		Usage:   "Organization management commands",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a new organization",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name",
						Usage: "Organization name",
					},
					cli.StringFlag{
						Name:  "description",
						Usage: "Organization description",
					},
				},
				Action: orgCreateAction,
			},
			{
				Name:   "list",
				Usage:  "List all organizations you belong to",
				Action: orgListAction,
			},
			{
				Name:  "members",
				Usage: "List organization members",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "org-id",
						Usage: "Organization ID",
					},
				},
				Action: orgMembersAction,
			},
			{
				Name:  "invite",
				Usage: "Invite a member to the organization",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "org-id",
						Usage: "Organization ID",
					},
					cli.StringFlag{
						Name:  "email",
						Usage: "Email address of user to invite",
					},
					cli.StringFlag{
						Name:  "role",
						Usage: "Role: admin or member",
						Value: "member",
					},
				},
				Action: orgInviteAction,
			},
		},
	}
}

func getAuthenticatedClient() (*api.Client, error) {
	// Load credentials
	creds, err := config.LoadCredentials()
	if err != nil {
		return nil, err
	}

	// Get API URL from environment or use default
	apiURL := os.Getenv("ENVV_API_URL")
	if apiURL == "" {
		apiURL = "https://api.envv.sh"
	}

	// Create authenticated client
	client := api.NewClient(apiURL)
	client.SetToken(creds.AccessToken)

	return client, nil
}

func orgCreateAction(c *cli.Context) error {
	// Get flags
	name := c.String("name")
	description := c.String("description")

	// Prompt for missing values
	if name == "" {
		fmt.Print("Organization name: ")
		fmt.Scanln(&name)
	}

	if name == "" {
		return fmt.Errorf("organization name is required")
	}

	// Get authenticated client
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	fmt.Println("🏢 Creating organization...")

	req := api.CreateOrgRequest{
		Name:        name,
		Description: description,
	}

	org, err := client.CreateOrganization(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	fmt.Printf("\n✅ Organization created successfully!\n")
	fmt.Printf("   ID:   %s\n", org.ID)
	fmt.Printf("   Name: %s\n", org.Name)
	fmt.Printf("   Slug: %s\n", org.Slug)
	fmt.Printf("   Role: %s\n", org.Role)

	return nil
}

func orgListAction(c *cli.Context) error {
	// Get authenticated client
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	orgs, err := client.ListOrganizations(ctx)
	if err != nil {
		return fmt.Errorf("failed to list organizations: %w", err)
	}

	if len(orgs) == 0 {
		fmt.Println("No organizations found. Create one with 'envv org create'")
		return nil
	}

	fmt.Printf("\n📋 Your Organizations (%d):\n\n", len(orgs))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSLUG\tROLE\tMEMBERS\tCREATED")
	fmt.Fprintln(w, "--\t----\t----\t----\t-------\t-------")

	for _, org := range orgs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
			org.ID, org.Name, org.Slug, org.Role, org.MemberCount, org.CreatedAt)
	}

	w.Flush()
	fmt.Println()

	return nil
}

func orgMembersAction(c *cli.Context) error {
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

	members, err := client.GetOrganizationMemberKeys(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization members: %w", err)
	}

	if len(members) == 0 {
		fmt.Println("No members found")
		return nil
	}

	fmt.Printf("\n👥 Organization Members (%d):\n\n", len(members))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EMAIL\tNAME\tROLE\tAGE PUBLIC KEY")
	fmt.Fprintln(w, "-----\t----\t----\t--------------")

	for _, member := range members {
		// Truncate public key for display
		pubKey := member.AgePublicKey
		if len(pubKey) > 20 {
			pubKey = pubKey[:20] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			member.Email, member.Name, member.Role, pubKey)
	}

	w.Flush()
	fmt.Println()

	return nil
}

func orgInviteAction(c *cli.Context) error {
	orgID := c.String("org-id")
	email := c.String("email")
	role := c.String("role")

	if orgID == "" {
		return fmt.Errorf("--org-id is required")
	}

	// Prompt for missing values
	if email == "" {
		fmt.Print("Email to invite: ")
		fmt.Scanln(&email)
	}

	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Validate role
	if role != "admin" && role != "member" {
		return fmt.Errorf("role must be 'admin' or 'member'")
	}

	// Get authenticated client
	client, err := getAuthenticatedClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	fmt.Printf("📧 Inviting %s as %s...\n", email, role)

	req := api.InviteMemberRequest{
		Email: email,
		Role:  role,
	}

	if err := client.InviteMember(ctx, orgID, req); err != nil {
		return fmt.Errorf("failed to invite member: %w", err)
	}

	fmt.Printf("✅ Invitation sent to %s\n", email)

	return nil
}
