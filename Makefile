PREFIX := /usr/local
SRCDIR :=

BINAME := yup

build:
	go build -v -o ${BINAME}

install:
	install -Dm755 ${BINAME} $(SRCDIR)$(PREFIX)/bin/${BINAME}

uninstall:
	rm -f $(SRCDIR)$(PREFIX)/bin/${BINAME}

test:
	go vet
	go test -v
