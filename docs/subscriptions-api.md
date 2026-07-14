# Subscriptions API 仕様・処理フロー

## 1. 概要

Subscriptions API は、fincodeを利用したカード決済サブスクリプションの開始、状態取得、解約を提供する。

サブスクリプション開始は非同期フローである。`POST /subscriptions/checkout` の成功時点では契約はまだ作成されず、カード登録用URLだけが発行される。利用者がfincode画面でカードを登録し、バックエンドが `customers.payment_methods.updated` Webhookでカードの有効化を確認した後にfincodeのサブスクリプションを作成する。

対象エンドポイントは以下のとおり。

| Method | Path | 用途 | 認証 |
| --- | --- | --- | --- |
| `GET` | `/subscriptions` | 現在の契約状態を取得 | Bearer JWT |
| `POST` | `/subscriptions/checkout` | カード登録URLを発行して契約開始フローへ進む | Bearer JWT |
| `DELETE` | `/subscriptions` | fincodeへ解約を要求 | Bearer JWT |
| `POST` | `/webhooks/fincode` | fincodeからカード登録・契約状態イベントを受信 | `Fincode-Signature` |

Swagger UIでは `subscriptions` タグに上3つ、`webhooks` タグにWebhookが表示される。

## 2. 共通仕様

### 2.1 ベースURL

ローカル環境の例：

```text
http://localhost:8080
```

### 2.2 認証

Subscriptions APIの3エンドポイントでは、次のAuthorizationヘッダーが必要となる。

```http
Authorization: Bearer <JWT>
```

JWTはHMAC署名で検証され、`user_id` claimが有効なUUIDである必要がある。未指定、不正な形式、署名不正、期限切れ、`user_id`不正の場合は `401 Unauthorized` を返す。

```text
unauthorized
```

`subsc` claimが `true` であることはSubscriptions APIの利用条件ではない。未契約ユーザーも開始APIや状態取得APIを利用できる。

### 2.3 エラーレスポンス

現在のエラーレスポンスはJSONではなく、`text/plain` の文字列である。

| HTTP status | 主な条件 | レスポンス例 |
| --- | --- | --- |
| `400 Bad Request` | IDや内部設定値のバリデーションエラー | `validation error` |
| `401 Unauthorized` | JWT不正、または有効なユーザーIDを取得できない | `unauthorized` |
| `404 Not Found` | ユーザーまたは契約IDが存在しない | `user not found` |
| `409 Conflict` | すでに `users.subsc = true` | `resource already exists` |
| `502 Bad Gateway` | fincode通信失敗、想定外のDBエラーなど | エンドポイント固有の固定文言 |

fincodeのエラーレスポンス本文はサーバー内部のエラーには保持されるが、Subscriptions APIのクライアントへはそのまま返されない。

## 3. GET /subscriptions

認証中ユーザーについて、ローカルDBに保存されている契約状態を返す。fincode APIへの問い合わせは行わない。

### リクエスト

```http
GET /subscriptions HTTP/1.1
Authorization: Bearer <JWT>
```

リクエストボディはない。

### 成功レスポンス

`200 OK`

```json
{
  "active": true,
  "fincode_customer_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "fincode_subscription_id": "subscription-id"
}
```

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `active` | boolean | `users.subsc` の現在値 |
| `fincode_customer_id` | string | fincode顧客ID。未作成の場合はフィールド自体を省略 |
| `fincode_subscription_id` | string | fincode契約ID。未作成の場合はフィールド自体を省略 |

未契約・fincode顧客未作成の場合：

```json
{
  "active": false
}
```

## 4. POST /subscriptions/checkout

fincodeのカード登録URLを発行する。このAPIはカード登録URLを作るだけであり、サブスクリプション契約そのものは作成しない。

### リクエスト

```http
POST /subscriptions/checkout HTTP/1.1
Authorization: Bearer <JWT>
```

リクエストボディはない。

### 成功レスポンス

`201 Created`

```json
{
  "checkout_url": "https://secure.test.fincode.jp/v1/links_customers_and_cards/...",
  "expires_at": "2026-07-14T12:07:23+09:00"
}
```

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `checkout_url` | string | 利用者を遷移させるfincodeカード登録URL |
| `expires_at` | RFC3339 datetime | URLの有効期限。バックエンドのAPIレスポンスではGoの `time.Time` としてJSON化される |

### 内部処理

1. JWTからユーザーIDを取得する。
2. `users` テーブルからユーザーを取得する。
3. `users.subsc = true` の場合は `409 Conflict` を返す。
4. `fincode_customer_id` が保存済みなら、その顧客IDを再利用する。
5. 顧客IDがない場合、fincode `POST /v1/customers` を呼び出す。
6. 作成した顧客IDを `users.fincode_customer_id` に保存する。
7. fincode `POST /v1/card_sessions` を呼び出す。
8. fincodeから返されたカード登録URLと有効期限をクライアントへ返す。

### fincode顧客作成

fincodeへ次の内容を送る。

```json
{
  "id": "<ユーザーUUID>",
  "email": "<ユーザーのメールアドレス>"
}
```

`idempotent_key` ヘッダーと顧客IDにはユーザーUUIDを使用する。

### fincodeカード登録URL作成

現在の設定ではURL有効期間は30分である。

```json
{
  "expire": "2026/07/14 12:07:23",
  "guide_mail_send_flag": "0",
  "completion_mail_send_flag": "0",
  "customer_id": "<fincode顧客ID>"
}
```

- `expire` はJSTへ変換し、fincode仕様の `yyyy/MM/dd HH:mm:ss` で送信する。
- ミリ秒、`Z`、`+09:00` はfincodeへ送る文字列には含めない。
- `shop_service_name` は現在空文字で初期化されており、`omitempty` により送信されない。
- fincodeからの案内メール・完了メールはどちらも無効である。
- 成功URL、キャンセルURLは現在送信していない。

## 5. カード登録後の契約作成フロー

利用者が `checkout_url` でカード登録を完了すると、fincodeから `customers.payment_methods.updated` Webhookを受信する。リダイレクト型カード登録の結果を受けるには、fincode側でこのイベントのWebhook設定が必要となる。

```text
クライアント
  │ POST /subscriptions/checkout
  ▼
バックエンド
  │ 必要ならfincode顧客を作成
  │ POST /v1/card_sessions
  ▼
カード登録URLを返却
  │
  ▼
利用者がfincode画面でカード登録
  │
  ▼
fincode ── customers.payment_methods.updated ──> バックエンド
  │                              │
  │                              ├─ POST /v1/subscriptions
  │                              └─ usersの契約情報を更新
  ▼                              ▼
サブスクリプション作成       active状態をDBへ保存
```

### `customers.payment_methods.updated` 受信時の処理

Webhook例：

```json
{
  "event": "customers.payment_methods.updated",
  "customer_id": "<fincode顧客ID>",
  "card_id": "<カードID>",
  "card_status": "ACTIVATED",
  "pay_type": "Card"
}
```

処理内容：

1. `pay_type = Card` かつ `card_status = ACTIVATED` であることを確認する。
2. 条件を満たさない通知は契約を作成せず、正常受信として返す。
3. `customer_id` からローカルユーザーを特定する。
4. すでに有効な契約IDを持っている場合は何もせず成功とする。
5. fincode `POST /v1/subscriptions` を呼び出す。
6. fincodeのレスポンスに含まれる契約IDと状態をDBへ保存する。

直接カードAPIを利用する既存フローとの互換性のため、`card.regist` も引き続き同じ契約作成処理の起点として受け付ける。

fincodeへ送る主な値：

```json
{
  "pay_type": "Card",
  "plan_id": "<FINCODE_PLAN_ID>",
  "customer_id": "<fincode顧客ID>",
  "card_id": "<カードID>",
  "start_date": "2026/07/14"
}
```

契約作成では `customer_id` と `card_id` を元に生成したSHA-256値を `idempotent_key` として送る。

契約ステータスが `ACTIVE` または `RUNNING` の場合は `users.subsc = true`、`CANCELED` または `INCOMPLETE` の場合は `false` として保存する。未知のステータスは契約作成直後には `false` として保存される。

### 契約状態Webhook

| event | 処理 |
| --- | --- |
| `customers.payment_methods.updated` | カードが `ACTIVATED` になったらサブスクリプションを作成 |
| `card.regist` | 直接カード登録API向けの互換フローとしてサブスクリプションを作成 |
| `subscription.card.regist` | 契約IDを保存し、ステータスに応じて `subsc` を更新 |
| `subscription.card.update` | 契約IDと `subsc` を更新 |
| `subscription.card.delete` | `subsc = false` に更新 |
| `subscription.card.cancel` | `subsc = false` に更新 |

`subscription.card.regist/update` で認識するstatusは `ACTIVE`、`RUNNING`、`CANCELED`、`INCOMPLETE` のみである。未知のstatusはWebhook処理エラーとなる。

## 6. DELETE /subscriptions

認証中ユーザーのfincodeサブスクリプションを解約する。

### リクエスト

```http
DELETE /subscriptions HTTP/1.1
Authorization: Bearer <JWT>
```

リクエストボディはない。

### 成功レスポンス

`202 Accepted`

レスポンスボディはない。

### 内部処理

1. JWTからユーザーIDを取得する。
2. ユーザーの `fincode_subscription_id` を取得する。
3. 契約IDがない場合は `404 Not Found` を返す。
4. fincodeへ次のリクエストを送る。

```http
DELETE /v1/subscriptions/{subscription_id}?pay_type=Card
```

5. fincodeへのリクエストが成功したら `202 Accepted` を返す。
6. この時点では `users.subsc` を変更しない。
7. 後続の `subscription.card.delete` または `subscription.card.cancel` Webhookで `users.subsc = false` にする。

したがって、解約APIの直後に `GET /subscriptions` を呼ぶと、Webhookが到着するまでは `active: true` が返る可能性がある。

## 7. Webhook仕様

### エンドポイント

```http
POST /webhooks/fincode
Fincode-Signature: <FINCODE_WEBHOOK_SIGNATURE>
Content-Type: application/json
```

WebhookはBearer JWTを使用しない。リクエストの `Fincode-Signature` と環境変数 `FINCODE_WEBHOOK_SIGNATURE` を定数時間比較し、一致した場合のみ処理する。

現在の実装はリクエストボディからHMACを計算する方式ではなく、設定済みの固定文字列との完全一致である。

リクエストボディの最大サイズは1 MiB。

### 成功レスポンス

認識対象イベントを正常処理した場合：

```json
{
  "receive": "0"
}
```

未対応イベントは再送を避けるため、処理せず `200 OK` を返す。未対応イベント時のレスポンスボディはない。

### エラー

| status | 条件 |
| --- | --- |
| `400` | JSON不正などペイロードをデコードできない |
| `401` | 署名未設定、ヘッダー未指定、署名不一致 |
| `405` | POST以外。ただし現在のルーターではPOSTパターンのみ登録 |
| `500` | 対応イベントのDB更新またはfincode API呼び出しに失敗 |

## 8. DB上の状態

契約状態は `users` テーブルの次の3列で管理する。

| 列 | 型 | 説明 |
| --- | --- | --- |
| `subsc` | boolean, default false | アプリ内で利用する契約有効フラグ |
| `fincode_customer_id` | varchar, unique, nullable | fincode顧客ID |
| `fincode_subscription_id` | varchar, unique, nullable | fincode契約ID |

解約Webhookでは `subsc` だけを `false` にし、`fincode_subscription_id` は削除しない。

JWTにも `subsc` claimが含まれるため、WebhookでDBが更新されても発行済みJWTのclaimは変化しない。ログイン処理は新しいJWTを発行するため、プレミアム機能へ契約状態を反映するには再ログインなどによるJWT更新が必要になる。

## 9. fincode通信共通仕様

バックエンドからfincodeへ送る共通ヘッダー：

```http
Authorization: Bearer <FINCODE_PRIVATE_KEY>
Api-Version: 20211001
Content-Type: application/json;charset=UTF-8
```

冪等キーがある処理では次も送信する。

```http
idempotent_key: <key>
```

- HTTPクライアントのタイムアウトは10秒。
- fincodeレスポンスの読み込み上限は1 MiB。
- 2xx以外はfincode APIエラーとして扱う。
- APIバージョンは設定がなければ `20211001`。

## 10. 必須環境変数

| 環境変数 | 用途 | 例 |
| --- | --- | --- |
| `FINCODE_API_BASE_URL` | fincode APIのベースURL | `https://api.test.fincode.jp` |
| `FINCODE_PRIVATE_KEY` | fincode Secret API Key | テストまたは本番の秘密鍵 |
| `FINCODE_PLAN_ID` | カード登録後に作成する契約のプランID | fincode上のプランID |
| `FINCODE_WEBHOOK_SIGNATURE` | Webhook署名照合用の固定値 | fincode側設定と同じ値 |

## 11. クライアント実装時の推奨フロー

1. `GET /subscriptions` で現在状態を確認する。
2. `active = false` なら `POST /subscriptions/checkout` を呼ぶ。
3. 受け取った `checkout_url` へ利用者を遷移させる。
4. カード登録完了後、バックエンドがWebhookを処理するまで待つ。
5. `GET /subscriptions` を再取得し、`active = true` を確認する。
6. 新しい契約状態を含むJWTが必要な画面では再ログインまたはトークン再発行を行う。
7. 解約時は `DELETE /subscriptions` を呼び、Webhook処理後に `active = false` になることを確認する。

## 12. 現在の実装上の注意点

- checkout成功は契約成功を意味しない。カード登録とWebhook処理が完了して初めて契約が有効になる。
- `GET /subscriptions` はfincodeの最新状態ではなくローカルDBを参照する。
- 解約はWebhook到着までローカル状態へ反映されない。
- checkout開始判定は `users.subsc` を使用する。`subsc = false` で古い契約IDだけが残っている場合でも、新しいカード登録URLを発行できる。
- Webhook失敗時は `500` を返すため、fincode側の再送により同じイベントが再度届く可能性がある。
- 顧客作成と契約作成には冪等キーがあるが、カード登録URL作成には冪等キーを付けていない。
- `shop_service_name`、成功URL、キャンセルURL、fincodeメール送信は現在未設定または無効である。
