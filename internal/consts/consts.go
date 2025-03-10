package consts

const (
	// Transaction Types
	Deposit    = "deposit"
	Withdrawal = "withdrawal"

	// Status types
	Pending    = "pending"
	Completed  = "completed"
	Processing = "processing"
)

const (
	DepositRoute  = "/deposit"
	WithdrawRoute = "/withdraw"
	CallbackRoute = "/callback"
	HealthRoute   = "/health"
)
