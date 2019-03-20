SOURCE_FILES := $(shell find . -type f -not -path './vendor/*' -iname '*.go')

LDFLAGS = -s -w -extldflags "-static"

.PHONY: all
all: build

.PHONY: build
build: update-gitignore

.PHONY: debug
debug:
	@$(MAKE) LDFLAGS= update-gitignore

.PHONY: format
format:
	goreturns -b -i -w $(SOURCE_FILES)

update-gitignore: $(SOURCE_FILES)
	go build -a -ldflags '$(LDFLAGS)'

.PHONY: install
install: all
	install -d $(DESTDIR)/usr/bin
	cp update-gitignore $(DESTDIR)/usr/bin
