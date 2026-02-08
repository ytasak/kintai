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
go build -o kn
```

## 使い方

サブコマンド・フラグ・値はすべて短縮形が使える（長い形式もそのまま使用可能）。

### 短縮マッピング

| 対象 | 長い形式 | 短縮形 |
|---|---|---|
| サブコマンド | `start` / `end` | `s` / `e` |
| フラグ | `--mode` / `--only` | `-m` / `-o` |
| mode値 | `office` / `remote` | `o` / `r` |
| only値 | `kinnosuke` / `slack` | `kin` / `s` |

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
go run . s -m o
go run . s -m r
go run . e

# 個別テスト（短縮形）
go run . s -m o -o kin
go run . e -o s

# 長い形式も使用可能
go run . start --mode office --only kinnosuke
go run . end --only slack
```

## ライセンス

Private
