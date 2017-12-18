build:
	go build -o app *.go

clean:
	@rm -rf app
	@rm -rf .glide
	@rm -rf vendor
	@rm -rf glide.lock

run: build
	./app

fmt:
	@gofmt -w *.go
	@gofmt -w **/**.go

deps:
	@curl https://glide.sh/get | sh
	@glide install