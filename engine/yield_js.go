//go:build js && wasm

package engine

import "runtime"

var yieldCounter int

// MaybeYield periodically yields to the browser event loop during search
// to prevent UI freezes. Called at the top of Negamax.
func MaybeYield() {
	yieldCounter++
	if yieldCounter&0x1FFF == 0 { // every ~8192 nodes
		runtime.Gosched()
	}
}
