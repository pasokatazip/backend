# PETYO-YO Backend

PETYO-YO のバックエンドです。Go 製の REST API、PostgreSQL、投稿内容を解析する Python ワーカー、定期処理用の cron プロセスで構成されています。

API 仕様は、起動後に [Swagger UI](http://localhost:8080/docs/) から確認できます。

## 使用技術

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Python](https://img.shields.io/badge/Python-3.12-3776AB?logo=python&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![pgvector](https://img.shields.io/badge/pgvector-384次元-4169E1)

| 分類 | 技術 | 用途 |
| --- | --- | --- |
| API | Go 1.26 / `net/http` | REST API、認証、ユースケースの実行 |
| データベース | PostgreSQL 16 / pgvector | 永続化、ベクトル検索、`LISTEN / NOTIFY` |
| DB ドライバー | pgx | Go から PostgreSQL へ接続 |
| マイグレーション | Goose | スキーマとマスターデータの管理 |
| Python ワーカー | Python 3.12 | 投稿イベントの購読、自然言語処理 |
| NLP / ML | Sentence Transformers / PyTorch / SudachiPy | 埋め込み生成、名詞抽出、グループ判定 |
| API ドキュメント | Swagger / Swaggo | OpenAPI ドキュメントの生成・表示 |
| 開発環境 | Docker Compose / Air | コンテナ構築、Go のホットリロード |

## アーキテクチャ

Go API はクリーンアーキテクチャをベースに、外側の層から内側の層へ依存する構成です。

```text
HTTP Request
    ↓
Router / Middleware
    ↓
Controller
    ↓
Usecase
    ↓
Domain（Repository interface）
    ↑
Infrastructure（PostgreSQL・fincode・Web Push）
```

### ディレクトリ構造

```text
.
├── cmd/
│   ├── server/               # REST API のエントリーポイント
│   └── cron/                 # 定期シミュレーション・通知処理
├── internal/
│   ├── controllers/          # HTTP 入出力、DTO
│   ├── domain/               # エンティティ、ドメインエラー、Repository interface
│   ├── usecases/             # アプリケーションのユースケース
│   ├── infrastructure/
│   │   ├── auth/             # JWT、パスワードハッシュ
│   │   ├── database/         # PostgreSQL 接続
│   │   ├── fincode/          # 決済 API クライアント
│   │   ├── middleware/       # 認証、CORS、サブスクリプション判定
│   │   ├── notification/     # Web Push
│   │   └── persistence/      # Repository の PostgreSQL 実装
│   ├── presenter/            # レスポンス形式への変換
│   ├── router/               # ルーティング
│   └── timeutil/             # 日時処理
├── python/
│   ├── listeners/            # PostgreSQL イベントリスナー
│   └── tasks/                # NLP・ペット行動処理
├── migrations/               # Goose マイグレーション
├── seeds/                    # 環境別シード
├── db/init/                  # DB 初期化 SQL
├── docs/                     # Swagger / API 補足資料
├── Dockerfile
└── docker-compose.yaml
```

レイヤーの責務は次のとおりです。

- `domain`: ビジネスルールと抽象（Repository interface）を保持します。
- `usecases`: ドメインを組み合わせ、アプリケーション固有の処理を実行します。
- `controllers` / `presenter`: HTTP とアプリケーション間のデータを変換します。
- `infrastructure`: DB、認証、決済、通知など外部システムとの接続を実装します。

## サードパーティ

| サービス・ライブラリ | 用途 | 設定 |
| --- | --- | --- |
| fincode | 単発決済・サブスクリプション決済 | `FINCODE_*` |
| Web Push / VAPID | ブラウザへの通知 | `VAPID_*` |
| JWT | API 認証トークン | `JWT_SECRET`, `JWT_EXP_MIN` |
| Google UUID | UUID の生成・処理 | アプリケーション内部 |
| bcrypt | パスワードのハッシュ化 | アプリケーション内部 |
| Swagger UI / Swaggo | API ドキュメント | `/docs/` |
| pgvector | 投稿・名詞・グループ名の埋め込み | PostgreSQL 拡張 |
| Sentence Transformers | 384 次元の文章埋め込み生成 | Python ワーカー |
| SudachiPy | 日本語の形態素解析・名詞抽出 | Python ワーカー |

## テーブル

| テーブル | 概要 |
| --- | --- |
| `users` | ユーザー、認証情報、課金状態 |
| `notifications` | ユーザー単位の通知設定と Web Push 購読情報 |
| `user_souvenir_praise_flags` | ユーザー・レポート対象日ごとの「ほめる」選択済み状態 |
| `pets` | ペットのプロフィール、特性値、現在の進化・グループ |
| `user_active_pets` | ユーザーと現在有効なペットの関連 |
| `pet_experiences` | ペットの累積経験値と給餌回数 |
| `pet_experience_events` | 経験値の獲得履歴 |
| `experience_caps` | 経験値獲得上限のマスタ |
| `evolution_stages` | 進化段階のマスタ |
| `evolution_rules` | 進化条件と遷移先 |
| `pet_evolutions` | ペットごとの進化履歴 |
| `posts` | ペットの投稿とベクトルデータ |
| `group_masters` | 興味グループと特性変化のマスタ |
| `group_keywords` | グループ判定に使うキーワード |
| `extracted_nouns` | 投稿から抽出した名詞とベクトル |
| `noun_group_matches` | 抽出名詞とグループのマッチ結果 |
| `pet_group_interests` | ペットごとのグループ興味スコア |
| `reports` | 時間単位の行動レポート |
| `pet_group_joins` | ペットのグループ参加・離脱履歴 |
| `pet_hourly_logs` | 時間単位のシミュレーションログ |
| `pet_interest_propagations` | ペット間で伝播した興味の履歴（受信ペットごとにJST日次2回まで） |
| `user_rumor_receipts` | ユーザーが受け取った群れの噂と元投稿の履歴 |
| `souvenir_masters` | おみやげのマスタ |
| `pet_souvenirs` | ペットが獲得したおみやげ |
| `pet_departure_rules` | ペットの旅立ち条件マスタ |
| `pet_departures` | ペットの旅立ち予定・状態 |

正確なカラム、制約、インデックスは [`migrations/`](./migrations/) を参照してください。

## 環境構築手順

### 前提

- Git
- Docker Desktop（Docker Compose v2 を含む）
- fincode のテスト用シークレットキー
- cron の通知処理も起動する場合は VAPID キー

Go、Python、PostgreSQL をホストへ個別にインストールする必要はありません。

### 1. リポジトリを取得する

```bash
git clone <repository-url>
cd backend-1
```

### 2. 環境変数を作成する

macOS / Linux:

```bash
cp .env.example .env
```

Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

`.env` の最低限必要な値を設定します。

```dotenv
POSTGRES_DB=hi
POSTGRES_USER=hi
POSTGRES_PASSWORD=hi
DATABASE_URL=postgres://hi:hi@postgres:5432/hi?sslmode=disable

PYTHON_SERVICE_URL=http://python:8000

JWT_SECRET=<十分に長いランダム文字列>
JWT_EXP_MIN=2880
CORS_ALLOWED_ORIGINS=http://localhost:3000

FINCODE_PRIVATE_KEY=<fincodeのテスト用シークレットキー>
FINCODE_API_BASE_URL=https://api.test.fincode.jp
FINCODE_BILLING_MODE=one_time
FINCODE_PURCHASE_SUCCESS_URL=http://localhost:3000/Subscription
FINCODE_PAYMENT_AMOUNT=999
```

`FINCODE_BILLING_MODE=subscription` を利用する場合は、`FINCODE_PLAN_ID` と `FINCODE_WEBHOOK_SIGNATURE` も設定してください。

cron コンテナを起動する場合は、さらに次を設定します。

```dotenv
VAPID_PUBLIC_KEY=<公開鍵>
VAPID_PRIVATE_KEY=<秘密鍵>
VAPID_SUBJECT=mailto:<連絡先メールアドレス>
```

### 3. DB を起動してマイグレーションする

```bash
docker compose up -d postgres
docker compose run --rm migrate
```

マイグレーションの状態は次のコマンドで確認できます。

```bash
docker compose run --rm migrate sh -c 'goose -dir migrations postgres "$DATABASE_URL" status'
```

### 4. API と Python ワーカーを起動する

```bash
docker compose up --build backend python
```

バックグラウンドで起動する場合:

```bash
docker compose up -d --build backend python
```

cron も含めてすべて起動する場合:

```bash
docker compose up -d --build
```

### 5. 動作確認

```bash
curl http://localhost:8080/health
```

`ok` が返れば API は起動しています。

| 対象 | URL |
| --- | --- |
| API | `http://localhost:8080` |
| ヘルスチェック | `http://localhost:8080/health` |
| Swagger UI | `http://localhost:8080/docs/` |
| PostgreSQL | `localhost:5432` |

ログの確認:

```bash
docker compose logs -f backend python
```

停止:

```bash
docker compose down
```

DB データも削除して初期化する場合のみ、次を実行します。

```bash
docker compose down -v
```

このコマンドは PostgreSQL の Docker ボリュームを削除するため、保存済みデータは復元できません。

## テスト

```bash
docker compose run --rm backend go test ./...
```
