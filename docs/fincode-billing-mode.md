# fincode課金方式の切り替え

## 設定

`FINCODE_BILLING_MODE`で起動時に課金方式を選択する。

| 値 | 課金方式 | 必須設定 | API |
| --- | --- | --- | --- |
| `one_time` | 買い切り（デフォルト） | なし | `POST /purchases/checkout`, `POST /purchases/confirm` |
| `subscription` | サブスクリプション | `FINCODE_PLAN_ID`, `FINCODE_WEBHOOK_SIGNATURE` | `POST /subscriptions/checkout`, `GET /subscriptions`, `DELETE /subscriptions` |

買い切り価格は`FINCODE_PAYMENT_AMOUNT`で指定する。未設定時は999円。未設定時のカード登録成功URLは`http://localhost:3000/Subscription`で、`FINCODE_PURCHASE_SUCCESS_URL`を設定すると変更できる。

## 買い切りフロー

1. フロントエンドが認証付きで`POST /purchases/checkout`を呼ぶ。
2. レスポンスの`checkout_url`へ遷移し、fincode画面でカードを登録する。
3. 登録完了後、fincodeが`FINCODE_PURCHASE_SUCCESS_URL`へPOSTリダイレクトする。
4. リダイレクト先のフロントエンドが、同じユーザーのJWTを付けて`POST /purchases/confirm`を呼ぶ。
5. バックエンドが登録カードを取得し、999円の一括決済を同期実行する。
6. `CAPTURED`を確認後、決済IDをDBへ保存し、`users.subsc=true`に更新する。
7. APIは`{"subsc":true}`を返す。決済IDは`users.fincode_subscription_id`へ保存するが、このレスポンスには含めない。

`POST /purchases/confirm`は再実行可能で、すでに購入済みのユーザーには再課金せず現在の購入状態を返す。

買い切りモードではWebhookを決済開始やDB更新に使用しない。カード登録イベントが届いても受領するだけで、処理は行わない。サブスクリプション用Webhook処理は`subscription`モード向けに残す。

## レイヤー構成

- `internal/usecases/onetime`: 買い切り固有ユースケース
- `internal/usecases/subsc`: サブスクリプション固有ユースケース
- `internal/usecases/ensure_fincode_customer.go`: 顧客作成・取得の共通処理
- `internal/infrastructure/fincode/card_session.go`: カード登録URL発行
- `internal/infrastructure/fincode/card.go`: 登録カード取得
- `internal/infrastructure/fincode/payment.go`: 一括決済
- `internal/infrastructure/fincode/subscription.go`: サブスクリプション

既存の`users.subsc`を有料権限フラグとして継続利用する。買い切りモードでは一度決済が成功すると永続的に`true`となる。既存DBとの互換性維持のため、`fincode_subscription_id`カラムには買い切り時の決済IDを保存する。
