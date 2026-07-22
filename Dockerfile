# ********************
# 開発用
# ********************
FROM golang:1.26-alpine AS dev

WORKDIR /app

# git: Go module取得
#ca-certificates: HTTPS通信
#air: Goホットリロード用
RUN apk add --no-cache git ca-certificates && \
    go install github.com/air-verse/air@v1.61.7 && \
    go install github.com/pressly/goose/v3/cmd/goose@latest

# go.modを先にコピーしてDocker cacheを効かせる
COPY go.mod ./

# go.sumがある場合だけコピー
COPY go.sum* ./

# Go moduleを取得
RUN go mod download

# アプリコピー
COPY . .

EXPOSE 8080

# airでホットリロード起動
CMD ["air"]

# ********************
# 本番ビルド
# ********************
FROM golang:1.26-alpine AS builder

WORKDIR /app

# ビルドに必要なパッケージ
RUN apk add --no-cache git ca-certificates

COPY go.mod ./
COPY go.sum* ./

RUN go mod download

# migration実行用のgooseをビルド
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

# 軽量なLinuxバイナリを作成
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cron ./cmd/cron

# ********************
# 本番実行
# ********************
FROM alpine:3.20 AS prod

WORKDIR /app

# HTTPS通信、JST のスケジューリング、DB 操作用に必要なパッケージ
RUN apk add --no-cache ca-certificates postgresql16-client tzdata

# builderから実行バイナリとmigration用gooseをコピー
COPY --from=builder /app/server ./server
COPY --from=builder /app/cron ./cron
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# migrationファイルを本番imageにも含める
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/seeds ./seeds
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint
RUN chmod +x /usr/local/bin/docker-entrypoint

EXPOSE 8080

# Go API と定期実行プロセスを起動
CMD ["/usr/local/bin/docker-entrypoint"]
