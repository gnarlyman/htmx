.PHONY: dev build generate

dev:
	air

build: generate
	go build -o tmp/app

generate: css templ

css:
	npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css --minify

templ:
	templ generate

watch:
	air -c .air.toml