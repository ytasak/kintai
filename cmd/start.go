package cmd

import (
	"context"
	"fmt"
	"os"

	"kintai/internal/kinnosuke"
	"kintai/internal/slackkintai"

	"github.com/spf13/cobra"
)

var startMode string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "出社打刻して、Slackの業務開始スレにリアクションする",
	RunE: func(cmd *cobra.Command, args []string) error {
		if startMode != "office" && startMode != "remote" {
			return fmt.Errorf("--mode must be office or remote")
		}

		// 1) 勤怠ノ助：出社
		t, err := kinnosuke.StampStart()
		if err != nil {
			return err
		}
		fmt.Printf("✔ 出社完了 (%s)\n", t)

		// 2) Slack：開始スレにリアクション
		if err := slackkintai.ReactStart(context.Background(), startMode); err != nil {
			return err
		}
		fmt.Println("✔ Slackリアクション完了 (開始)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVar(&startMode, "mode", "", "office|remote (required)")
	_ = startCmd.MarkFlagRequired("mode")

	// envの未設定を早めに気づけるように（任意）
	startCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		_ = os.Getenv("KIN_COMPANYCD")
		_ = os.Getenv("KIN_LOGINCD")
		_ = os.Getenv("KIN_PASSWORD")
		_ = os.Getenv("SLACK_TOKEN")
		_ = os.Getenv("SLACK_CHANNEL")
		return nil
	}
}
