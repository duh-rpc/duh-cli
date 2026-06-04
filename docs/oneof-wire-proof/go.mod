// Standalone reference harness — NOT part of the duh-cli module. It is built and
// run manually (see README.md) against protoc-generated code in a scratch dir.
// Keeping it a separate module excludes it from the duh-cli `go test ./...` build.
module proof

go 1.24
