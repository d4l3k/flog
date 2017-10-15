ARGS?=

.PHONY: build
build:
	go install -v .

.PHONY: run
run: build
	flog $(ARGS)
