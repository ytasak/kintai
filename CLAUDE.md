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

# Slack OAuth（kn auth 用）
SLACK_CLIENT_ID="..."
SLACK_CLIENT_SECRET="..."
```

| 変数名 | 用途 |
|---|---|
| `KIN_COMPANYCD` | 勤之助の会社コード |
| `KIN_LOGINCD` | 勤之助のログインID |
| `KIN_PASSWORD` | 勤之助のパスワード |
| `SLACK_TOKEN` | Slackトークン（User Token `xoxp-...` 推奨。`kn auth` で自動取得可能） |
| `SLACK_CHANNEL` | Slackチャンネル名またはID（`C`始まりのID推奨） |
| `SLACK_CLIENT_ID` | Slack AppのClient ID（`kn auth` 用） |
| `SLACK_CLIENT_SECRET` | Slack AppのClient Secret（`kn auth` 用） |

### Slackトークンの種類

| トークン | プレフィックス | リアクション主体 |
|---|---|---|
| User Token | `xoxp-...` | 本人（推奨） |
| Bot Token | `xoxb-...` | アプリ（Bot） |

必要なスコープ: `reactions:write`, `channels:history`, `channels:read`

## CLIコマンド

バイナリ名は `kn`。サブコマンド・フラグ・値はすべて短縮形が使える（長い形式もそのまま使用可能）。

### 短縮マッピング

| 対象 | 長い形式 | 短縮形 |
|---|---|---|
| バイナリ | `kintai` | `kn` |
| サブコマンド | `start` / `end` / `auth` | `s` / `e` / `a` |
| フラグ | `--mode` / `--only` | `-m` / `-o` |
| mode値 | `office` / `remote` | `o` / `r` |
| only値 | `kinnosuke` / `slack` | `kin` / `s` |

### `kn start` (`kn s`) - 出社打刻

```bash
kn s -m <o|r> [-o <kin|s>]
# 長い形式: kn start --mode <office|remote> [--only <kinnosuke|slack>]
```

| フラグ | 必須 | 値 | 説明 |
|---|---|---|---|
| `-m` / `--mode` | Yes | `o`(office) / `r`(remote) | 出社種別 |
| `-o` / `--only` | No | `kin`(kinnosuke) / `s`(slack) | 片方だけ実行（省略時は両方） |

```bash
# 出社（勤之助 + Slack）
kn s -m o
kn s -m r

# 勤之助のみ
kn s -m o -o kin

# Slackのみ
kn s -m o -o s
```

### `kn end` (`kn e`) - 退社打刻

```bash
kn e [-o <kin|s>]
# 長い形式: kn end [--only <kinnosuke|slack>]
```

| フラグ | 必須 | 値 | 説明 |
|---|---|---|---|
| `-o` / `--only` | No | `kin`(kinnosuke) / `s`(slack) | 片方だけ実行（省略時は両方） |

```bash
# 退社（勤之助 + Slack）
kn e

# 勤之助のみ
kn e -o kin

# Slackのみ
kn e -o s
```

### `kn auth` (`kn a`) - Slack OAuth トークン取得

```bash
kn a
# 長い形式: kn auth
```

Slack OAuth 2.0 フローを実行し、User Token (`xoxp-...`) を取得して `.env` に自動保存する。

**前提条件:**
1. Slack App の **Redirect URLs** に `http://localhost:9876/callback` を追加
2. **User Token Scopes** に `reactions:write`, `channels:history`, `channels:read` を追加
3. `.env` に `SLACK_CLIENT_ID` と `SLACK_CLIENT_SECRET` を記載

**処理フロー:**
1. ローカルサーバー (`localhost:9876`) を起動
2. ブラウザで Slack 認可ページを開く
3. ユーザーが認可すると、コールバックで認可コードを受信
4. コード → トークン交換
5. 取得した User Token を `.env` の `SLACK_TOKEN` に保存

## ビルド・実行

```bash
# 依存解決
go mod tidy

# ビルドして実行
go build -o kn
./kn s -m o
./kn e

# ビルドせずに直接実行（開発時）
go run . s -m o
go run . e
```

## アーキテクチャ

```
main.go          → エントリポイント（godotenvで.env読み込み → cmd.Execute()）
cmd/
  root.go        → Cobra CLIのルートコマンド定義（バイナリ名: kn）
  start.go       → `kn start(s)` 出社コマンド（-m, -o フラグ、値の短縮正規化）
  end.go         → `kn end(e)` 退社コマンド（-o フラグ、値の短縮正規化）
  auth.go        → `kn auth(a)` Slack OAuthトークン取得コマンド
internal/
  auth/
    oauth.go     → Slack OAuth 2.0 フロー（ローカルサーバー・ブラウザ認可・トークン交換）
    dotenv.go    → .envファイル更新ユーティリティ（SLACK_TOKENの上書き・追加）
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
