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

// normalizeMode は --mode の短縮値を正規化する
func normalizeMode(v string) string {
	switch v {
	case "o":
		return "office"
	case "r":
		return "remote"
	default:
		return v
	}
}

// normalizeOnly は --only の短縮値を正規化する
func normalizeOnly(v string) string {
	switch v {
	case "kin":
		return "kinnosuke"
	case "s":
		return "slack"
	default:
		return v
	}
}

var startCmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"s"},
	Short:   "出社打刻して、Slackの業務開始スレにリアクションする",
	RunE: func(cmd *cobra.Command, args []string) error {
		startMode = normalizeMode(startMode)
		if startMode != "office" && startMode != "remote" {
			return fmt.Errorf("--mode(-m) must be office(o) or remote(r)")
		}

		// --only バリデーション
		if startOnly != "" && startOnly != "kinnosuke" && startOnly != "slack" {
			return fmt.Errorf("--only(-o) must be kinnosuke(kin) or slack(s)")
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
	startCmd.Flags().StringVarP(&startMode, "mode", "m", "", "office(o)|remote(r) (required)")
	startCmd.Flags().StringVarP(&startOnly, "only", "o", "", "kinnosuke(kin)|slack(s) (省略時は両方実行)")
	_ = startCmd.MarkFlagRequired("mode")

	// envの未設定を早めに気づけるように（任意）
	startCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		startOnly = normalizeOnly(startOnly)
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
