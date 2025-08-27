# Meteora ☄️ DbcGOSDK

## Overview

Meteora Dynamic Bonding Curve program in Go. Powered by [solana-go](https://github.com/gagliardetto/solana-go) and [solana-anchor-go](https://github.com/fragmetric-labs/solana-anchor-go).

## Running test...

Tests are located in ./helpers/ and ./maths

- to run all test in a directory:

 > go test ./helpers/

- to run a single test file:

> go test ./helpers/curve_test.go

- to run a specific test func:

> go test ./helpers/ -run "TestBuildCurve"

- run a specific sub-test:

> go test ./helpers/ -run "TestBuildCurve/xxxxx"

- run all test:
> make test


## Examples