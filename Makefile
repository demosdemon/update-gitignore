SOURCE_FILES := $(shell find . -type f -not -path './vendor/*' -iname '*.go')

.PHONY: format
format:
	goreturns -b -i -w $(SOURCE_FILES)
