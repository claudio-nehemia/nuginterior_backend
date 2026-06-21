package constants

type ContextKey string

const (
	ContextKeyUserEmail      ContextKey = "user_email"
	ContextKeyCompanyID      ContextKey = "company_id"
	ContextKeyUserRole       ContextKey = "user_role"
	ContextKeyFilterCompanyID ContextKey = "filter_company_id"
)
