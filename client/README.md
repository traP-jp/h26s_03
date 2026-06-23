# client

## 最初にやること

リポジトリ直下で次を実行します。

```bash
pnpm --dir client install
```

## 開発

リポジトリ直下で次を実行します。

```bash
mise run client
```

ブラウザで `http://localhost:5173` を開きます。API はデフォルトで `http://localhost:8080` を見に行きます。

バックエンドと MySQL も一緒に起動したいときは、リポジトリ直下で次を使います。

```bash
mise run dev
```

## よく触る場所

- `src/App.vue`: アプリ全体の外枠
- `src/router.ts`: URL と画面の対応
- `src/views`: ページ単位の画面
- `src/components`: 使い回す UI 部品
- `src/lib/api.ts`: API クライアント
- `src/gen/api-types.ts`: OpenAPI から生成された型

## API を呼ぶ

API のレスポンス型は `src/gen/api-types.ts` に生成されます。API 定義を変えたら、リポジトリ直下で次を実行します。

```bash
mise run codegen
```

`src/gen/api-types.ts` は生成ファイルなので、基本的に手で編集しません。API の形を変えたいときは `../openapi/openapi.yaml` を編集してください。

## 変更後の確認

フロントエンドを触ったら、コミット前に次を実行します。

```bash
mise run client-fmt
mise run client-lint
mise run client-typecheck
mise run client-build
```

## 環境変数

API の接続先を変えたいときは `VITE_API_BASE` を指定します。

```bash
VITE_API_BASE=http://localhost:8080 pnpm --dir client run dev
```
