PREFIX := /usr/local
SRCDIR :=

BINAME := yup
PKGBUILD := yup

build:
	go get
	go install
	go build -v -o ${BINAME}

install:
	install -Dm755 ${BINAME} $(SRCDIR)$(PREFIX)/bin/${BINAME}
	install -Dm755 completions/zsh $(SRCDIR)$(PREFIX)/share/zsh/site-functions/completions/_${PKGBUILD}

uninstall:
	rm -f $(SRCDIR)$(PREFIX)/bin/${BINAME}

test:
	go vet
	go test -v ./...
