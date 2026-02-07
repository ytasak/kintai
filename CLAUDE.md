# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

Go製のCLIツール。勤怠管理システム「勤之助（Kinnosuke）」への打刻と、Slackチャンネルへのリアクション付与を1コマンドで行う。

## 必須環境変数

```bash
# 勤怠ノ助
export KIN_COMPANYCD="..."
export KIN_LOGINCD="..."
export KIN_PASSWORD="..."

# Slack
export SLACK_TOKEN="xoxb-..."
export SLACK_CHANNEL="mikasa-kintai"   # 可能ならCxxxxのID推奨
```

| 変数名 | 用途 |
|---|---|
| `KIN_COMPANYCD` | 勤之助の会社コード |
| `KIN_LOGINCD` | 勤之助のログインID |
| `KIN_PASSWORD` | 勤之助のパスワード |
| `SLACK_TOKEN` | Slack Botトークン（reactions:write, channels:history, channels:read権限が必要） |
| `SLACK_CHANNEL` | Slackチャンネル名またはID（`C`始まりのID推奨） |

## ビルド・実行

```bash
# 依存解決
go mod tidy

# ビルドして実行
go build
./kintai start --mode office
./kintai start --mode remote
./kintai end

# ビルドせずに直接実行（開発時）
go run . start --mode office
go run . start --mode remote
go run . end
```

## アーキテクチャ

```
main.go          → cmd.Execute() を呼ぶだけのエントリポイント
cmd/
  root.go        → Cobra CLIのルートコマンド定義
  start.go       → `kintai start --mode <office|remote>` 出社コマンド
  end.go         → `kintai end` 退社コマンド
internal/
  kinnosuke/
    client.go    → 勤之助へのHTTPクライアント（Cookie管理・フォームPOST）
    parse.go     → HTMLのregexパース、ログイン、CSRF取得、打刻処理
  slackkintai/
    slack.go     → Slackチャンネルから当日のリマインダーメッセージを検索し、リアクション付与
```

### 処理フロー（start / end共通）

1. 環境変数からKinnosuke認証情報を読み込み
2. HTTPクライアント作成（Cookie jarでセッション管理）
3. 勤之助トップページをGETし、ログイン済みか確認（`<div class="user_name">`の有無）
4. 未ログインならPOSTでログイン
5. HTMLからCSRFトークンをregexで抽出
6. 打刻リクエストをPOST（type=1が出社、type=2が退社）
7. 再度トップページをGETし、打刻時刻をHTMLから抽出して表示
8. Slackで当日の該当リマインダーメッセージ（例: `リマインダー：業務開始スレ`）を検索し、リアクション付与

### Slackリアクション

- 出社(office): `:shussha:`
- 出社(remote): `:remote-start:`
- 退社: `:tai-kin:`

## 技術スタック

- Go 1.25.4
- CLI: `github.com/spf13/cobra`
- Slack: `github.com/slack-go/slack`
- 勤之助連携: 標準`net/http` + regexによるHTMLスクレイピング（外部パーサー不使用）
