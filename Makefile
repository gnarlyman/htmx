.PHONY: dev templ tailwind server

dev:
	make -j3 templ tailwind server

templ:
	templ generate --watch --proxy=http://localhost:8080 --proxyport=7331 --open-browser=false

tailwind:
	npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css --watch

server:
	reflex -r '\.go$$' -s -- go run main.go