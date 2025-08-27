## CONTRIBUTION GUIDE

This project is a direct implementation of the client-facing SDK of the [Meteora ☄️](https://www.meteora.ag/) [DBC JS/TS SDK](https://github.com/MeteoraAg/dynamic-bonding-curve-sdk) in the [Go programming language](https://go.dev/). Contributions are welcome, and here are some guidelines and tasks to note.

### Guides
- The project currently uses [solana-go](https://github.com/gagliardetto/solana-go) as its version of solana-web3js. No alternative libraries of this kind will be accepted, to ensure type compatibility and avoid conflicts.  

- The project also uses [solana-anchor-go](https://github.com/fragmetric-labs/solana-anchor-go) from Fragmetric Labs to generate interactions with the smart contract from the IDL, located in the `./generated` directory. There are Makefile targets for this, and the dependency is included in the `go.mod` file via `go tools`.  

- For now, until advised otherwise by the Meteora dev team, this library will closely mirror the JS/TS SDK — variable names, function calls, and logic flow will remain consistent.  

- For tests involving blockchain interaction, we use [Surfpool’s Surfnet](https://docs.surfpool.run/rpc/surfnet). Version [v0.10.3](https://github.com/txtx/surfpool/releases/tag/v0.10.3) is currently recommended.  

- Idiomatic Go code is recommended wherever possible.  

### Tasks

- Resolve handling of optional accounts for `swap` & `swap2`, specifically with `referralTokenAccount`. This issue lies outside the codebase itself; the related issue is raised [here](https://github.com/fragmetric-labs/solana-anchor-go/issues/20).  

- Validation functions present in the JS/TS SDK are not yet implemented here. Search the repo for `TODO` comments to find these areas.  

- More tests are needed. If you’re writing a test that does not exist in the JS/TS SDK, please also provide a control example in this [repo](https://github.com/tobyweb3x/dbcGoSDK) under the branch `tobytobias.sol`, or create your own branch if necessary. A control test here means implementing the same test with the JS/TS SDK as well, so the results can be directly compared.

- Benchmark tests, coverage tests, and code optimizations are all highly encouraged.  