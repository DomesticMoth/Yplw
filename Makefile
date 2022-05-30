clear:
	rm -rf ./out

build:
	go build -mod vendor -o out/native/ypw src/*.go

build-x64:
	env GOOS=linux GOARCH=amd64 go build -mod vendor -o out/x64/ypw src/*.go

build-arm7:
	env GOOS=linux GOARCH=arm GOARM=7 go build -mod vendor -o out/arm7/ypw src/*.go
