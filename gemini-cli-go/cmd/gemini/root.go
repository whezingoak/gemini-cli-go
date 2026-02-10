package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whezingoak/gemini-cli-go/pkg/commands"
	"github.com/whezingoak/gemini-cli-go/pkg/config"
	"github.com/whezingoak/gemini-cli-go/pkg/terminal"
)

var rootCmd = &cobra.Command{
	Use:   "gemini",
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
