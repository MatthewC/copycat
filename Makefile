ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif


build:
	go build -o bin/copycat main.go colors.go commands.go util.go files.go

run:
	go run main.go colors.go commands.go util.go files.go default

help:
	go run main.go colors.go commands.go util.go files.go help

list:
	go run main.go colors.go commands.go util.go files.go list

upload:
	go run main.go colors.go commands.go util.go files.go upload sample

download:
	go run main.go colors.go commands.go util.go files.go download sample

configure:
	go run main.go colors.go commands.go util.go files.go configure

files:
	go run main.go colors.go commands.go util.go files.go files

install:
	install -d $(DESTDIR)$(PREFIX)/lib/
	install -m 755 bin/copycat $(DESTDIR)$(PREFIX)/bin/