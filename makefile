build:
	GOARCH=wasm GOOS=js go build -o web/app.wasm
	go build

run: build
	./aspiratv

sass:
	sass  frontend/sass/mystyles.scss:web/mystyles.css