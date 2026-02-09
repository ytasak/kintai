# kintai

勤怠管理システム「勤之助（Kinnosuke）」への打刻と、Slackチャンネルへのリアクション付与を1コマンドで行うCLIツール。

## できること

- 勤之助への出社・退社打刻
- Slackの勤怠リマインダーメッセージへのリアクション自動付与
- `--only` フラグで勤之助・Slackを個別に実行可能
- Slack OAuth 2.0 による User Token の自動取得（`kn auth`）

## 必要なもの

- Go 1.25.4+
- 勤之助アカウント
- Slack App（`reactions:write`, `channels:history`, `channels:read` 権限）

### Slackトークンについて

本ツールは **User Token**（`xoxp-...`）を使用します。本人としてリアクションが付与されます。
`kn auth` コマンドで OAuth フローを実行すれば自動取得できます。

## セットアップ

### 1. 環境変数の設定

プロジェクトルートに `.env` ファイルを作成するか、シェルの環境変数として設定してください。

```bash
# 勤怠ノ助
KIN_COMPANYCD="..."
KIN_LOGINCD="..."
KIN_PASSWORD="..."

# Slack
SLACK_TOKEN="xoxp-..."             # kn auth で自動取得可能
SLACK_CHANNEL="..."      # CxxxxのチャンネルID推奨

# Slack OAuth（kn auth 用）
SLACK_CLIENT_ID="..."
SLACK_CLIENT_SECRET="..."
```

> `SLACK_TOKEN` は `kn auth` コマンドで自動取得・保存できます。手動設定も可能です。

### 2. Slack App の設定（`kn auth` を使う場合）

1. [api.slack.com/apps](https://api.slack.com/apps) で App を作成（または既存の App を使用）
2. **Basic Information** → App Credentials から `Client ID` / `Client Secret` を `.env` に記載
3. **OAuth & Permissions** → **Redirect URLs** に `https://localhost:9876/callback` を追加
4. **OAuth & Permissions** → **User Token Scopes** に以下を追加:
   - `reactions:write`
   - `channels:history`
   - `channels:read`

### 3. ビルド

```bash
go mod tidy
go build -o kn
```

## 使い方

サブコマンド・フラグ・値はすべて短縮形が使える（長い形式もそのまま使用可能）。

### 短縮マッピング

| 対象 | 長い形式 | 短縮形 |
|---|---|---|
| サブコマンド | `start` / `end` / `auth` | `s` / `e` / `a` |
| フラグ | `--mode` / `--only` | `-m` / `-o` |
| mode値 | `office` / `remote` | `o` / `r` |
| only値 | `kinnosuke` / `slack` | `kin` / `s` |

### Slack認証 (`auth` / `a`)

```bash
kn a
# 長い形式: kn auth
```

Slack OAuth 2.0 フローを実行し、User Token を取得して `.env` に自動保存します。

1. ローカルに HTTPS サーバーを起動
2. ブラウザで Slack 認可ページを開く
3. 認可後、取得したトークンを `.env` の `SLACK_TOKEN` に保存

> 初回のコールバック時にブラウザが証明書の警告を出す場合があります。「詳細設定」→「localhost にアクセスする」で続行してください。

### 出社打刻 (`start` / `s`)

```bash
kn s -m <o|r> [-o <kin|s>]
# 長い形式: kn start --mode <office|remote> [--only <kinnosuke|slack>]
```

| フラグ | 必須 | 値 | 説明 |
|---|---|---|---|
| `-m` / `--mode` | Yes | `o`(office) / `r`(remote) | 出社種別 |
| `-o` / `--only` | No | `kin`(kinnosuke) / `s`(slack) | 片方だけ実行（省略時は両方） |

```bash
# 出社（オフィス）- 勤之助 + Slack
./kn s -m o

# 出社（リモート）- 勤之助 + Slack
./kn s -m r

# 勤之助のみ
./kn s -m o -o kin

# Slackのみ
./kn s -m r -o s
```

### 退社打刻 (`end` / `e`)

```bash
kn e [-o <kin|s>]
# 長い形式: kn end [--only <kinnosuke|slack>]
```

| フラグ | 必須 | 値 | 説明 |
|---|---|---|---|
| `-o` / `--only` | No | `kin`(kinnosuke) / `s`(slack) | 片方だけ実行（省略時は両方） |

```bash
# 退社 - 勤之助 + Slack
./kn e

# 勤之助のみ
./kn e -o kin

# Slackのみ
./kn e -o s
```

### 出力例

```
✔ 出社完了 (09:00)
✔ Slackリアクション完了 (開始)
```

## Slackリアクション

| コマンド | リアクション |
|---|---|
| `s -m o` (`start --mode office`) | :shussha: |
| `s -m r` (`start --mode remote`) | :remote-start: |
| `e` (`end`) | :tai-kin: |

Slackチャンネル内の当日のリマインダーメッセージ（`リマインダー：業務開始スレ` / `リマインダー：業務終了スレ`）を自動検索し、リアクションを付与します。

## プロジェクト構成

```
main.go              エントリポイント（.env読み込み）
cmd/
  root.go            Cobra CLIルートコマンド
  start.go           出社コマンド (kn start / kn s)
  end.go             退社コマンド (kn end / kn e)
  auth.go            Slack認証コマンド (kn auth / kn a)
internal/
  auth/
    oauth.go         Slack OAuth 2.0 フロー（HTTPS・ブラウザ認可・トークン交換）
    dotenv.go        .envファイル更新ユーティリティ
  kinnosuke/
    client.go        勤之助HTTPクライアント（Cookie/セッション管理）
    parse.go         HTMLパース・ログイン・CSRF取得・打刻処理
  slackkintai/
    slack.go         Slackリアクション付与
```

## 開発

```bash
# ビルドせずに直接実行
go run . s -m o
go run . s -m r
go run . e

# Slack認証
go run . a

# 個別テスト（短縮形）
go run . s -m o -o kin
go run . e -o s

# 長い形式も使用可能
go run . start --mode office --only kinnosuke
go run . end --only slack
```

## ライセンス

Private
