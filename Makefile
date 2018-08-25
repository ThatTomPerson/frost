GO = go1.11

FUNCTION_NAME ?= $(notdir $(abspath .))
PKGS = $(shell $(GO) list -f '{{.Dir}}' ./...)
SRC = $(addsuffix /*.go,$(PKGS))

run: $(FUNCTION_NAME)
	cd tests; ../$<

$(FUNCTION_NAME): $(SRC)
	@$(GO) build -o $(FUNCTION_NAME)

test:
	@$(go) test ./...