# kintai

勤怠管理システム「勤之助（Kinnosuke）」への打刻と、Slackチャンネルへのリアクション付与を1コマンドで行うCLIツール。

## できること

- 勤之助への出社・退社打刻
- Slackの勤怠リマインダーメッセージへのリアクション自動付与
- `--only` フラグで勤之助・Slackを個別に実行可能

## 必要なもの

- Go 1.25.4+
- 勤之助アカウント
- Slackトークン（`reactions:write`, `channels:history`, `channels:read` 権限）

### Slackトークンについて

| トークン | プレフィックス | リアクション主体 |
|---|---|---|
| User Token | `xoxp-...` | 本人（推奨） |
| Bot Token | `xoxb-...` | アプリ（Bot） |

本人としてリアクションしたい場合は **User Token**（`xoxp-...`）を使用してください。

## セットアップ

### 1. 環境変数の設定

プロジェクトルートに `.env` ファイルを作成するか、シェルの環境変数として設定してください。

```bash
# 勤怠ノ助
KIN_COMPANYCD="..."
KIN_LOGINCD="..."
KIN_PASSWORD="..."

# Slack
SLACK_TOKEN="xoxp-..."
SLACK_CHANNEL="mikasa-kintai"   # CxxxxのチャンネルID推奨
```

### 2. ビルド

```bash
go mod tidy
go build
```

## 使い方

### 出社打刻 (`start`)

```bash
kintai start --mode <office|remote> [--only <kinnosuke|slack>]
```

| フラグ | 必須 | 値 | 説明 |
|---|---|---|---|
| `--mode` | Yes | `office` / `remote` | 出社種別 |
| `--only` | No | `kinnosuke` / `slack` | 片方だけ実行（省略時は両方） |

```bash
# 出社（オフィス）- 勤之助 + Slack
./kintai start --mode office

# 出社（リモート）- 勤之助 + Slack
./kintai start --mode remote

# 勤之助のみ
./kintai start --mode office --only kinnosuke

# Slackのみ
./kintai start --mode remote --only slack
```

### 退社打刻 (`end`)

```bash
kintai end [--only <kinnosuke|slack>]
```

| フラグ | 必須 | 値 | 説明 |
|---|---|---|---|
| `--only` | No | `kinnosuke` / `slack` | 片方だけ実行（省略時は両方） |

```bash
# 退社 - 勤之助 + Slack
./kintai end

# 勤之助のみ
./kintai end --only kinnosuke

# Slackのみ
./kintai end --only slack
```

### 出力例

```
✔ 出社完了 (09:00)
✔ Slackリアクション完了 (開始)
```

## Slackリアクション

| コマンド | リアクション |
|---|---|
| `start --mode office` | :shussha: |
| `start --mode remote` | :remote-start: |
| `end` | :tai-kin: |

Slackチャンネル内の当日のリマインダーメッセージ（`リマインダー：業務開始スレ` / `リマインダー：業務終了スレ`）を自動検索し、リアクションを付与します。

## プロジェクト構成

```
main.go              エントリポイント（.env読み込み）
cmd/
  root.go            Cobra CLIルートコマンド
  start.go           出社コマンド (kintai start)
  end.go             退社コマンド (kintai end)
internal/
  kinnosuke/
    client.go        勤之助HTTPクライアント（Cookie/セッション管理）
    parse.go         HTMLパース・ログイン・CSRF取得・打刻処理
  slackkintai/
    slack.go         Slackリアクション付与
```

## 開発

```bash
# ビルドせずに直接実行
go run . start --mode office
go run . start --mode remote
go run . end

# 個別テスト
go run . start --mode office --only kinnosuke
go run . end --only slack
```

## ライセンス

Private
