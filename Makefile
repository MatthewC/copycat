ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif


build:
	go build -o bin/copycat main.go colors.go commands.go util.go
run:
	go run main.go colors.go commands.go util.go default

help:
	go run main.go colors.go commands.go util.go help

list:
	go run main.go colors.go commands.go util.go list

upload:
	go run main.go colors.go commands.go util.go upload sample

download:
	go run main.go colors.go commands.go util.go download sample

configure:
	go run main.go colors.go commands.go util.go configure

install:
	install -d $(DESTDIR)$(PREFIX)/lib/
	install -m 755 bin/copycat $(DESTDIR)$(PREFIX)/bin/