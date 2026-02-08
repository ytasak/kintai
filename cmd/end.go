package cmd

import (
	"context"
	"fmt"
	"os"

	"kintai/internal/kinnosuke"
	"kintai/internal/slackkintai"

	"github.com/spf13/cobra"
)

var endOnly string

var endCmd = &cobra.Command{
	Use:     "end",
	Aliases: []string{"e"},
	Short:   "退社打刻して、Slackの業務終了スレにリアクションする",
	RunE: func(cmd *cobra.Command, args []string) error {
		// --only バリデーション
		if endOnly != "" && endOnly != "kinnosuke" && endOnly != "slack" {
			return fmt.Errorf("--only(-o) must be kinnosuke(kin) or slack(s)")
		}

		// 勤怠ノ助：退社
		if endOnly == "" || endOnly == "kinnosuke" {
			t, err := kinnosuke.StampEnd()
			if err != nil {
				return err
			}
			fmt.Printf("✔ 退社完了 (%s)\n", t)
		}

		// Slack：終了スレにリアクション
		if endOnly == "" || endOnly == "slack" {
			if err := slackkintai.ReactEnd(context.Background()); err != nil {
				return err
			}
			fmt.Println("✔ Slackリアクション完了 (終了)")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(endCmd)
	endCmd.Flags().StringVarP(&endOnly, "only", "o", "", "kinnosuke(kin)|slack(s) (省略時は両方実行)")

	// envの未設定を早めに気づけるように（任意）
	endCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		endOnly = normalizeOnly(endOnly)
		if endOnly == "" || endOnly == "kinnosuke" {
			_ = os.Getenv("KIN_COMPANYCD")
			_ = os.Getenv("KIN_LOGINCD")
			_ = os.Getenv("KIN_PASSWORD")
		}
		if endOnly == "" || endOnly == "slack" {
			_ = os.Getenv("SLACK_TOKEN")
			_ = os.Getenv("SLACK_CHANNEL")
		}
		return nil
	}
}
