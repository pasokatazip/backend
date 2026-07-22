#!/bin/sh
set -eu

./server &
server_pid=$!

./cron &
cron_pid=$!

shutdown() {
    kill -TERM "$server_pid" "$cron_pid" 2>/dev/null || true
}

trap shutdown INT TERM

# API または定期実行プロセスが終了したら、もう一方も停止してコンテナを終了する。
set +e
wait -n "$server_pid" "$cron_pid"
status=$?
set -e

shutdown
wait "$server_pid" 2>/dev/null || true
wait "$cron_pid" 2>/dev/null || true

exit "$status"
