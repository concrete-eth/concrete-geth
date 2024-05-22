# Concrete

Concrete is a framework for building application-specific rollups on the OP Stack.

Concrete blockchains are general-purpose EVM rollups that have been _enhanced_ at the protocol level to meet the needs of a specific use case, like expensive cryptography, spatial indexing, or complex digital physics.

With the Concrete API, you can build a rollup that fits your needs, without having to deal with the low-level complexity of forking an Ethereum client.

With Concrete, you can:

- Write app-specific code in any language that compiles to WASM.

- Add supercharged, stateful precompiles to the EVM using common structures like maps, arrays, and structs.

- Use existing EVM-compatible tools like MUD, Foundry, Otterscan, and more.

## Installing the Concrete CLI

### Requirements:

- Go 1.21 or higher
- `$GOPATH/bin` or `$GOBIN` in `$PATH` ([more info](https://go.dev/doc/code#Command))

### Installation:

Run the following command from the project root to install the `concrete` CLI tool:

```bash
go install ./concrete/cmd/concrete
```

