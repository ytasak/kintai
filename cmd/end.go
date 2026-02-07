package cmd

import (
	"context"
	"fmt"

	"kintai/internal/kinnosuke"
	"kintai/internal/slackkintai"

	"github.com/spf13/cobra"
)

var endCmd = &cobra.Command{
	Use:   "end",
	Short: "退社打刻して、Slackの業務終了スレにリアクションする",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1) 勤怠ノ助：退社
		t, err := kinnosuke.StampEnd()
		if err != nil {
			return err
		}
		fmt.Printf("✔ 退社完了 (%s)\n", t)

		// 2) Slack：終了スレにリアクション
		if err := slackkintai.ReactEnd(context.Background()); err != nil {
			return err
		}
		fmt.Println("✔ Slackリアクション完了 (終了)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(endCmd)
}
