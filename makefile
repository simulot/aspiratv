build:
	GOARCH=wasm GOOS=js go build -o web/app.wasm
	go build

run: build
	./aspiratv

sass:
	sass  frontend/sass/mystyles.scss:web/mystyles.css

clean:
	rm ./tmp/*


init:
	npm init
	npm install node-sass --save-dev
	npm install bulma --save-dev
	make sass