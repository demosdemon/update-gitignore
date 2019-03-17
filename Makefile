SOURCE_FILES := $(shell find . -type f -not -path './vendor/*' -iname '*.go')
PACKAGE := github.com/demosdemon/update-gitignore

os = darwin linux openbsd windows
arch = 386 amd64
builds = $(foreach goos,$(os),$(foreach goarch,$(arch),$(goos)/$(goarch)))

.PHONY: help
help:
	@echo 'builds = $(builds)'

.PHONY: format
format:
	goreturns -b -i -w $(SOURCE_FILES)
