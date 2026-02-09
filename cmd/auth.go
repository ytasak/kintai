package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"kintai/internal/auth"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:     "auth",
	Aliases: []string{"a"},
	Short:   "Slack OAuth 2.0 で User Token を取得し .env に保存する",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID := os.Getenv("SLACK_CLIENT_ID")
		clientSecret := os.Getenv("SLACK_CLIENT_SECRET")

		if clientID == "" || clientSecret == "" {
			return fmt.Errorf("SLACK_CLIENT_ID と SLACK_CLIENT_SECRET を .env またはシェル環境変数に設定してください")
		}

		token, err := auth.Run(context.Background(), clientID, clientSecret)
		if err != nil {
			return err
		}

		// .env にトークンを保存
		envPath := ".env"
		if err := auth.UpsertEnvToken(envPath, "SLACK_TOKEN", token); err != nil {
			return fmt.Errorf(".envへの書き込みに失敗: %w", err)
		}

		masked := maskToken(token)
		fmt.Printf("✔ SLACK_TOKEN を .env に保存しました (%s)\n", masked)
		return nil
	},
}

// maskToken はトークンの先頭10文字だけ表示し、残りをマスクする。
func maskToken(token string) string {
	if len(token) <= 10 {
		return strings.Repeat("*", len(token))
	}
	return token[:10] + strings.Repeat("*", len(token)-10)
}

func init() {
	rootCmd.AddCommand(authCmd)
}
