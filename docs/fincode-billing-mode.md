# fincode 課金方式の切替

## 設定

`FINCODE_BILLING_MODE` で起動時に課金方式を選択する。

| 値 | 課金方式 | 必須設定 | API |
| --- | --- | --- | --- |
| `one_time` | 買い切り（デフォルト） | なし | `POST /purchases/checkout`, `GET /purchases` |
| `subscription` | 既存サブスクリプション | `FINCODE_PLAN_ID` | `POST /subscriptions/checkout`, `GET /subscriptions`, `DELETE /subscriptions` |

買い切り価格は `cmd/server/main.go` の `fincodePaymentAmount` に税込請求総額を日本円の整数で指定する。

## レイヤー構成

- `internal/usecases/onetime`: 買い切り固有のユースケース
- `internal/usecases/subsc`: 既存サブスクリプション固有のユースケース
- `internal/usecases/ensure_fincode_customer.go`: 顧客作成・取得の共通処理
- `internal/infrastructure/fincode/card_session.go`: カード登録URL発行の共通処理
- `internal/infrastructure/fincode/payment.go`: 買い切り決済のfincodeアダプター
- `internal/infrastructure/fincode/subscription.go`: サブスクリプションのfincodeアダプター

ドメイン層では顧客、カード登録セッション、サブスクリプション、買い切り決済の各Gatewayを分割している。各ユースケースは必要な小さいインターフェースだけに依存する。

## 買い切りフロー

1. クライアントが `POST /purchases/checkout` を呼ぶ。
2. 共通処理がfincode顧客を確保し、カード登録URLを返す。
3. カード登録完了後、fincodeから `customers.payment_methods.updated` Webhookを受け取る。
4. `POST /v1/payments` で `CAPTURE` 決済を登録する。
5. `PUT /v1/payments/{id}` で登録カードへの請求を実行する。
6. `CAPTURED` を確認して永続利用権を有効化する。

同じ顧客に対する決済IDと冪等性キーは決定的に生成し、利用権が有効なユーザーは再課金しない。

## 既存データとの互換性

既存の `users.subsc` は有料機能の利用権フラグとして継続利用する。買い切りモードでは一度決済が成功すると永続的に `true` となる。

既存DBを破壊せず切り戻せるよう、`fincode_subscription_id` カラムは課金プロバイダーIDの保存先として維持している。買い切りモードでは決済IDが入る。アプリケーションの買い切りコードからは汎用名 `FincodeBillingID` / `UpdateFincodeBilling` を使用し、レガシー名を境界の外へ漏らさない。
