ARGS?=

.PHONY: build
build:
	go build -v .

.PHONY: install
install:
	go install -v .

.PHONY: run
run: install
	flog $(ARGS)
