# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

NAME := gsync

# Set the build dir, where built cross-compiled binaries will be output
BUILDDIR := ${PREFIX}/bin

all: clean build

deps:
	dep ensure

.PHONY: build
build: $(NAME)

$(NAME): *.go
	mkdir -p $(BUILDDIR)
	@echo "+ $@"
	go build -o $(BUILDDIR)/$(NAME) .

.PHONY: clean
clean: ## Cleanup any build binaries or packages
	@echo "+ $@"
	$(RM) $(NAME)
	$(RM) -r $(BUILDDIR)
