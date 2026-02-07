package main

import (
	"kintai/cmd"

	"github.com/joho/godotenv"
)

func main() {
	// .envがあれば環境変数にロード（なくてもエラーにならない）
	godotenv.Load()
	cmd.Execute()
}
