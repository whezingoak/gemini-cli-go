package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/whezingoak/gemini-cli-go/pkg/api"
	"github.com/whezingoak/gemini-cli-go/pkg/terminal"
)

var ChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := api.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Start interactive UI
		return terminal.StartChat(client)
	},
}

func init() {
	// Add flags if needed
}
