//go:build !(js && wasm)

package engine

// MaybeYield is a no-op on native builds.
func MaybeYield() {}
