
dbc:
	@go tool solana-anchor-go -src=./idl/dynamic-bonding-curve/idl.json -pkg=dbc -dst=./generated/dbc/
dammv2:
	@go tool solana-anchor-go -src=./idl/dammv2/idl.json -pkg=dammv2 -dst=./generated/dammv2/
dammv1:
	@go tool solana-anchor-go -src=./idl/dammv1/idl.json -pkg=dammv1 -dst=./generated/dammv1/
dv:
	@go tool solana-anchor-go -src=./idl/dynamic-vault/idl.json -pkg=dynamic_vault -dst=./generated/dynamicVault/

test:
	@go test ./helpers/
	@go test ./maths/...
	@go test .