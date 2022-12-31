ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif

LD_FLAGS = -ldflags "-X main.VersionLog=$(VERSION_LOG) -X main.VersionHost=$(VERSION_HOST)"

build:
	go build ${LD_FLAGS} -o bin/copycat ./...

build-all:
	GOOS=linux GOARCH=amd64 go build ${LD_FLAGS} -o bin/copycat-linux-amd64 ./...
	GOOS=linux GOARCH=arm64 go build ${LD_FLAGS} -o bin/copycat-linux-arm64 ./...
	GOOS=linux GOARCH=386 go build ${LD_FLAGS} -o bin/copycat-linux-386 ./...
	GOOS=darwin GOARCH=amd64 go build ${LD_FLAGS} -o bin/copycat-darwin-amd64 ./...
	GOOS=darwin GOARCH=arm64 go build ${LD_FLAGS} -o bin/copycat-darwin-arm64 ./...
	# GOOS=windows GOARCH=amd64 go build ${LD_FLAGS} -o bin/copycat-windows-amd64.exe ./...
	# GOOS=windows GOARCH=arm64 go build ${LD_FLAGS} -o bin/copycat-windows-arm64.exe ./...
	# GOOS=windows GOARCH=386 go build ${LD_FLAGS} -o bin/copycat-windows-386.exe ./...

run:
	go run ./... $(CMD)

install:
	install -d $(DESTDIR)$(PREFIX)/lib/
	install -m 755 bin/copycat $(DESTDIR)$(PREFIX)/bin/

test:
	go test -v ./...

clean:
	rm -fr bin/
	rm CURRENT_VERSION