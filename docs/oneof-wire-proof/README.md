# oneOf wire-compatibility proof

This is the empirical proof that the **nested, key-tagged union** form is byte-for-byte
wire-compatible between protobuf and JSON — the basis for the `oneOf` style DUH supports.
See [`../oneof-wire-compatibility.html`](../oneof-wire-compatibility.html) for the full explainer.

## What it proves

Using the real protobuf toolchain (`protoc` + `protojson`, the exact marshaler `duh.go` uses),
it proves that a protobuf `oneof` and plain optional message fields produce **identical wire bytes**
and **interoperate in both directions** with each other and with a generic JSON consumer:

- proto `oneof` JSON  ≡  optional-fields JSON  ≡  `{"cat_event":{"pet_name":"Whiskers"}}` (nested, snake_case preserved)
- A server using either proto form is readable by a client using the other (JSON **and** binary wire).
- A non-protobuf JSON consumer (standing in for an OpenAPI-driven client) reads/writes the same shape.
- A real proto `oneof` additionally **rejects** two-variant input, enforcing "exactly one" — matching OpenAPI `oneOf` semantics.

## Re-running

```sh
cd "$(mktemp -d)"
cp /path/to/docs/oneof-wire-proof/{union.proto,proof_test.go} .
printf 'module proof\n\ngo 1.24\n' > go.mod
protoc --go_out=. --go_opt=module=proof union.proto
go mod tidy
go test -v ./...
```

Requires `protoc`, `protoc-gen-go`, and network access for `go mod tidy`.
