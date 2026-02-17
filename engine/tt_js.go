//go:build js && wasm

package engine

func init() {
	// Reduce transposition table from 16MB to 2MB in Wasm builds
	// to keep total memory usage reasonable in browsers.
	TT = NewTranTable(2)
}
