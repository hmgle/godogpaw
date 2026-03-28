.PHONY: wasm serve clean native

PORT ?= 8088

wasm: web/godogpaw.wasm web/wasm_exec.js

web/godogpaw.wasm: wasm/main.go engine/*.go
	GOOS=js GOARCH=wasm go build -o web/godogpaw.wasm ./wasm/

web/wasm_exec.js:
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/wasm_exec.js 2>/dev/null || \
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" web/wasm_exec.js

serve: wasm
	@echo "Serving at http://localhost:$(PORT)"
	cd web && python3 -m http.server $(PORT)

native:
	go build -o godogpaw .

clean:
	rm -f web/godogpaw.wasm web/wasm_exec.js godogpaw
