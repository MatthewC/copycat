ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif

build:
	go build -o bin/copycat *.go

build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/copycat-linux-amd64 *.go
	GOOS=linux GOARCH=arm64 go build -o bin/copycat-linux-arm64 *.go
	GOOS=linux GOARCH=386 go build -o bin/copycat-linux-386 *.go
	GOOS=darwin GOARCH=amd64 go build -o bin/copycat-darwin-amd64 *.go
	GOOS=darwin GOARCH=arm64 go build -o bin/copycat-darwin-arm64 *.go
	# GOOS=windows GOARCH=amd64 go build -o bin/copycat-windows-amd64.exe *.go
	# GOOS=windows GOARCH=arm64 go build -o bin/copycat-windows-arm64.exe *.go
	# GOOS=windows GOARCH=386 go build -o bin/copycat-windows-386.exe *.go
	go run *.go version-clean > CURRENT_VERSION
run:
	go run *.go $(CMD)

install:
	install -d $(DESTDIR)$(PREFIX)/lib/
	install -m 755 bin/copycat $(DESTDIR)$(PREFIX)/bin/

clean:
	rm -fr bin/
	rm CURRENT_VERSION