package domain

import "context"

// Transaction はコールバック内の処理を一括で確定または取り消す。
// リポジトリの呼び出しには、コールバックに渡されたコンテキストを使用する。
type Transaction interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}
