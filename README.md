# kintai

勤怠管理システム「勤之助（Kinnosuke）」への打刻と、Slackチャンネルへのリアクション付与を1コマンドで行うCLIツール。

## できること

- 勤之助への出社・退社打刻
- Slackの勤怠リマインダーメッセージへのリアクション自動付与

## 必要なもの

- Go 1.25.4+
- 勤之助アカウント
- Slack Botトークン（`reactions:write`, `channels:history`, `channels:read` 権限）

## セットアップ

### 1. 環境変数の設定

```bash
# 勤怠ノ助
export KIN_COMPANYCD="..."
export KIN_LOGINCD="..."
export KIN_PASSWORD="..."

# Slack
export SLACK_TOKEN="xoxb-..."
export SLACK_CHANNEL="mikasa-kintai"   # CxxxxのチャンネルID推奨
```

### 2. ビルド

```bash
go mod tidy
go build
```

## 使い方

```bash
# 出社（オフィス）
./kintai start --mode office

# 出社（リモート）
./kintai start --mode remote

# 退社
./kintai end
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
main.go              エントリポイント
cmd/
  root.go            Cobra CLIルートコマンド
  start.go           出社コマンド (kintai start --mode <office|remote>)
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
```

## ライセンス

Private
