package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/whezingoak/gemini-cli-go/pkg/api"
	"github.com/whezingoak/gemini-cli-go/pkg/commands"
	"github.com/whezingoak/gemini-cli-go/pkg/config"
	"github.com/whezingoak/gemini-cli-go/pkg/terminal"
)

var rootCmd = &cobra.Command{
	Use:   "gemini [prompt]",
	Short: "Gemini CLI - Interact with Google's Gemini models",
	Long:  `Gemini CLI is a tool to interact with Google's Gemini models directly from your terminal.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Init(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
		// Apply theme (could be loaded from config later)
		terminal.ApplyTheme(terminal.DefaultDark)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := api.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		if len(args) == 0 {
			// Interactive Chat Mode
			return terminal.StartChat(client)
		} else {
			// One-off Prompt Mode
			prompt := strings.Join(args, " ")
			resp, err := client.GenerateContent(ctx, prompt)
			if err != nil {
				return err
			}
			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				fmt.Println(resp.Candidates[0].Content.Parts[0].Text)
			}
			return nil
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("project", "", "Google Cloud Project ID")
	rootCmd.PersistentFlags().String("location", "us-central1", "API Location")
	rootCmd.PersistentFlags().String("model", "gemini-1.5-pro", "Model to use")

	rootCmd.AddCommand(commands.ChatCmd)
}
