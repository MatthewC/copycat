ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif

build:
	go build -o bin/copycat *.go

run:
	go run *.go $(CMD)

install:
	install -d $(DESTDIR)$(PREFIX)/lib/
	install -m 755 bin/copycat $(DESTDIR)$(PREFIX)/bin/

clean:
	rm -fr bin/