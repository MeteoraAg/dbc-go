package anchor

type Program[A PgAccountI, M PgMethodI] struct {
	Accounts PgAccounts[A]
	Methods  PgMethods[M]
	// Views
}
