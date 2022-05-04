# go-hasdefer

## what

a go linter to check that all goroutines have a defer statement.

## why

it's too easy to miss panics when they happen inside goroutines, since they exit the defer scope of the caller.

## install

`go install github.com/nathants/go-hasdefer@latest`


## usage

```bash

>> go-hasdefer $(find test/good/ -name '*.go')

>> go-hasdefer $(find test/bad/ -name '*.go')
missing defer anon func oneliner:        test/bad/bad_onliner.go:4      go func() {}()
missing defer anon func multiliner:      test/bad/bad_multiliner.go:4   go func() {
missing defer top level func multiliner: test/bad/bad_import.go:12 func Foobar() {
missing defer top level func oneliner:   test/bad/bad_import.go:6 func (d *Data) Foobar2() {}
missing defer named func multiliner:     test/bad/bad_imported.go:8     Foobar3 := func() {
missing defer top level func multiliner: test/bad/bad_import.go:8 func (d *Data) Foobar4() {
missing defer named func oneliner:       test/bad/bad_imported.go:12    Foobar4 := func() {}
```
