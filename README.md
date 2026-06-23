# h26s_03

traP 春ハッカソン 2026 3 班の Web アプリです。

チームメンバー:

- @cp20
- @renkon
- @nature36
- @Ayuto1123
- @msk❤️
- @azukimaru

## 最初にやること

このリポジトリは `client` が Vue、`server` が Go、DB が MySQL です。OpenAPI の定義をもとに、フロントエンドとバックエンドで使う型や handler を生成します。

1. Docker Desktop など、Docker が動く状態にする
2. `mise` を入れる
3. 依存関係を入れる

```bash
mise trust
mise install
mise run setup
```

`mise run setup` はフロントエンドの依存関係に加えて、VS Code の Go 整形で使う `goimports` も `.tools/bin/goimports` に入れます。

## 開発を始める

フロントエンド、バックエンド、MySQL をまとめて起動します。

```bash
mise run dev
```

起動できたら、ブラウザで `http://localhost:5173` を開きます。API は `http://localhost:8080` で動きます。

開発を止めるときは次を実行します。

```bash
mise run dev-down
```

## よく触る場所

### フロントエンド

バックエンドは `client/` 以下に全てまとまっています。

- 新しい画面を追加したいとき
  - `client/src/views` に `***View.vue` みたいなファイルを追加する
  - `router.ts` にいい感じに追加する

### バックエンド

バックエンドは `server/` 以下に全てまとまっています。

- 新しいエンドポイントを追加したい時
  - `server/internal/handlers`
- データベースの定義を変えたい時
  - TODO

## API を変更するとき

API の URL、リクエスト、レスポンスを変えたいときは、まず `openapi/openapi.yaml` を編集します。その後、次のコマンドで生成コードを更新します。

```bash
mise run codegen
```

生成されたファイルは主に次の場所に入ります。

- `client/src/gen/api-types.ts`
- `server/internal/gen/openapi/**`

生成後に `client` や `server` の実装で型エラーが出たら、API の変更に合わせて直してください。

## 作業前後に確認すること

フロントエンドを変更したら、次を確認します。

```bash
mise run client-fmt
mise run client-lint
mise run client-typecheck
mise run client-build
```

バックエンドを変更したら、次を確認します。

```bash
mise run server-build
mise run server-test
```

`server-test` は Docker を使って MySQL コンテナを立てるため、Docker が起動している必要があります。
