package database

import "database/sql"

// ExecOptionalTx lets mocked services exercise transactional code paths
// without requiring a real database handle.
func ExecOptionalTx(context *DbContext, fn func(*sql.Tx) error) error {
	if context == nil || context.GetDatabase() == nil {
		return fn(nil)
	}

	return context.ExecTx(fn)
}
