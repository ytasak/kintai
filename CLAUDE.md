# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

Go製のCLIツール。勤怠管理システム「勤之助（Kinnosuke）」への打刻と、Slackチャンネルへのリアクション付与を1コマンドで行う。

## 必須環境変数

`.env` ファイルに記載するか、シェルの環境変数として設定する（godotenvにより `.env` を自動読み込み）。

```bash
# 勤怠ノ助
KIN_COMPANYCD="..."
KIN_LOGINCD="..."
KIN_PASSWORD="..."

# Slack
SLACK_TOKEN="xoxp-..."
SLACK_CHANNEL="mikasa-kintai"   # 可能ならCxxxxのID推奨
```

| 変数名 | 用途 |
|---|---|
| `KIN_COMPANYCD` | 勤之助の会社コード |
| `KIN_LOGINCD` | 勤之助のログインID |
| `KIN_PASSWORD` | 勤之助のパスワード |
| `SLACK_TOKEN` | Slackトークン（User Token `xoxp-...` 推奨。本人としてリアクションされる） |
| `SLACK_CHANNEL` | Slackチャンネル名またはID（`C`始まりのID推奨） |

### Slackトークンの種類

| トークン | プレフィックス | リアクション主体 |
|---|---|---|
| User Token | `xoxp-...` | 本人（推奨） |
| Bot Token | `xoxb-...` | アプリ（Bot） |

必要なスコープ: `reactions:write`, `channels:history`, `channels:read`

## CLIコマンド

### `kintai start` - 出社打刻

```bash
kintai start --mode <office|remote> [--only <kinnosuke|slack>]
```

| フラグ | 必須 | 値 | 説明 |
|---|---|---|---|
| `--mode` | Yes | `office` / `remote` | 出社種別 |
| `--only` | No | `kinnosuke` / `slack` | 片方だけ実行（省略時は両方） |

```bash
# 出社（勤之助 + Slack）
kintai start --mode office
kintai start --mode remote

# 勤之助のみ
kintai start --mode office --only kinnosuke

# Slackのみ
kintai start --mode office --only slack
```

### `kintai end` - 退社打刻

```bash
kintai end [--only <kinnosuke|slack>]
```

| フラグ | 必須 | 値 | 説明 |
|---|---|---|---|
| `--only` | No | `kinnosuke` / `slack` | 片方だけ実行（省略時は両方） |

```bash
# 退社（勤之助 + Slack）
kintai end

# 勤之助のみ
kintai end --only kinnosuke

# Slackのみ
kintai end --only slack
```

## ビルド・実行

```bash
# 依存解決
go mod tidy

# ビルドして実行
go build
./kintai start --mode office
./kintai end

# ビルドせずに直接実行（開発時）
go run . start --mode office
go run . end
```

## アーキテクチャ

```
main.go          → エントリポイント（godotenvで.env読み込み → cmd.Execute()）
cmd/
  root.go        → Cobra CLIのルートコマンド定義
  start.go       → `kintai start` 出社コマンド（--mode, --only フラグ）
  end.go         → `kintai end` 退社コマンド（--only フラグ）
internal/
  kinnosuke/
    client.go    → 勤之助へのHTTPクライアント（Cookie管理・フォームPOST）
    parse.go     → HTMLのregexパース、ログイン、CSRF取得、打刻処理
  slackkintai/
    slack.go     → Slackチャンネルから当日のリマインダーメッセージを検索し、リアクション付与
```

### 処理フロー（start / end共通）

1. `.env` があれば環境変数にロード（godotenv）
2. `--only` の値に応じて勤之助・Slackの片方または両方を実行：

**勤之助の処理:**
1. 環境変数からKinnosuke認証情報を読み込み
2. HTTPクライアント作成（Cookie jarでセッション管理）
3. 勤之助トップページをGETし、ログイン済みか確認（`<div class="user_name">`の有無）
4. 未ログインならPOSTでログイン
5. HTMLからCSRFトークンをregexで抽出
6. 打刻リクエストをPOST（type=1が出社、type=2が退社）
7. 再度トップページをGETし、打刻時刻をHTMLから抽出して表示

**Slackの処理:**
1. 当日の該当リマインダーメッセージ（例: `リマインダー：業務開始スレ`）を検索
2. リアクション付与

### Slackリアクション

- 出社(office): `:shussha:`
- 出社(remote): `:remote-start:`
- 退社: `:tai-kin:`

## 技術スタック

- Go 1.25.4
- CLI: `github.com/spf13/cobra`
- Slack: `github.com/slack-go/slack`
- 環境変数: `github.com/joho/godotenv`
- 勤之助連携: 標準`net/http` + regexによるHTMLスクレイピング（外部パーサー不使用）
