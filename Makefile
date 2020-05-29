
CGO_ENABLED:=0
NAME:=proxysocket
GITCOMMIT:=$(shell git describe --always)
LDFlags:=-ldflags=" -X github.com/sharego/proxysocket/cmd.GitCommit=$(GITCOMMIT) -X github.com/sharego/proxysocket/lib.ForceCheckBoth=yes"

debug:
	CGO_ENABLED=$(CGO_ENABLED) go build $(LDFlags) -o $(NAME) .

.PHONY : all

all: clean binary

binary:
	@test -d bin || mkdir bin
	CGO_ENABLED=$(CGO_ENABLED) GOOS="linux" go build $(LDFlags) -o bin/$(NAME)-linux-amd64 .
	CGO_ENABLED=$(CGO_ENABLED) GOOS="darwin" go build $(LDFlags) -o bin/$(NAME)-darwin-amd64 .
	CGO_ENABLED=$(CGO_ENABLED) GOOS="windows" go build $(LDFlags) -o bin/$(NAME)-windows-amd64 .
	CGO_ENABLED=$(CGO_ENABLED) GOOS="linux" GOARCH="arm64" go build $(LDFlags) -o bin/$(NAME)-linux-aarch64 .

.PHONY : clean
clean:
	rm -f $(NAME) bin/$(NAME)-*
