package cmd

import (
	"context"
	"fmt"
	"os"

	"kintai/internal/kinnosuke"
	"kintai/internal/slackkintai"

	"github.com/spf13/cobra"
)

var (
	startMode string
	startOnly string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "出社打刻して、Slackの業務開始スレにリアクションする",
	RunE: func(cmd *cobra.Command, args []string) error {
		if startMode != "office" && startMode != "remote" {
			return fmt.Errorf("--mode must be office or remote")
		}

		// --only バリデーション
		if startOnly != "" && startOnly != "kinnosuke" && startOnly != "slack" {
			return fmt.Errorf("--only must be kinnosuke or slack")
		}

		// 勤怠ノ助：出社
		if startOnly == "" || startOnly == "kinnosuke" {
			t, err := kinnosuke.StampStart()
			if err != nil {
				return err
			}
			fmt.Printf("✔ 出社完了 (%s)\n", t)
		}

		// Slack：開始スレにリアクション
		if startOnly == "" || startOnly == "slack" {
			if err := slackkintai.ReactStart(context.Background(), startMode); err != nil {
				return err
			}
			fmt.Println("✔ Slackリアクション完了 (開始)")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVar(&startMode, "mode", "", "office|remote (required)")
	startCmd.Flags().StringVar(&startOnly, "only", "", "kinnosuke|slack (省略時は両方実行)")
	_ = startCmd.MarkFlagRequired("mode")

	// envの未設定を早めに気づけるように（任意）
	startCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if startOnly == "" || startOnly == "kinnosuke" {
			_ = os.Getenv("KIN_COMPANYCD")
			_ = os.Getenv("KIN_LOGINCD")
			_ = os.Getenv("KIN_PASSWORD")
		}
		if startOnly == "" || startOnly == "slack" {
			_ = os.Getenv("SLACK_TOKEN")
			_ = os.Getenv("SLACK_CHANNEL")
		}
		return nil
	}
}
