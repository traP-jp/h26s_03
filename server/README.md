# server

Go + Echo + MySQL で書くバックエンドです。API の処理、DB 初期化、migration はここに置きます。

## 開発

リポジトリ直下で次を実行します。

```bash
mise run server
```

Docker Compose で MySQL と server が起動します。API は `http://localhost:8080` で動きます。

フロントエンドも一緒に起動したいときは、リポジトリ直下で次を使います。

```bash
mise run dev
```

## よく触る場所

- `cmd/api/main.go`: サーバー起動処理、middleware、ルーティング
- `internal/handlers`: API の中身。1エンドポイント1ファイルで実装します。
- `internal/middleware/authx`: 認証関連 middleware
- `internal/gen/openapi`: OpenAPI から生成された handler interface など
- `migrations`: DB migration

## API を追加・変更する

1. `../openapi/openapi.yaml` に API の仕様を書く
2. リポジトリ直下で生成コードを更新する
3. `internal/handlers` に処理を書く

```bash
mise run codegen
```

`internal/gen/openapi` は生成ファイルなので、基本的に手で編集しません。

## DB を変更する

テーブルを変えたいときは `migrations` に SQL を追加・編集します。ローカル DB に反映するときは、リポジトリ直下で次を実行します。

```bash
mise run migrate-up
```

1 つ戻したいときは次を実行します。

```bash
mise run migrate-down
```

開発用の初期データを入れ直したいときは、server 起動後に次の API を呼びます。

```bash
curl -X POST http://localhost:8080/api/initialize
```

## 変更後の確認

バックエンドを触ったら、コミット前に次を実行します。

```bash
mise run server-build
mise run server-test
```

`server-test` は Docker を使って MySQL コンテナを立てます。Docker が起動していないと失敗します。

## 環境変数

- `API_ADDR` (default: `:8080`)
- `DB_DSN` (default: `app:app@tcp(localhost:3306)/app?parseTime=true&multiStatements=true`)
- `AUTH_MODE` (`SOFT` or `HARD`, default: `SOFT`)
- `ASSETS_DIR` (指定時は静的配信)
