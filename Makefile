
CGO_ENABLED:=0
NAME:=proxysocket

debug:
	CGO_ENABLED=$(CGO_ENABLED) go build -o $(NAME) .

.PHONY : all

all: clean binary

binary:
	@test -d bin || mkdir bin
	CGO_ENABLED=$(CGO_ENABLED) GOOS="linux" go build -o bin/$(NAME)-linux-amd64 .
	CGO_ENABLED=$(CGO_ENABLED) GOOS="darwin" go build -o bin/$(NAME)-darwin-amd64 .
	CGO_ENABLED=$(CGO_ENABLED) GOOS="windows" go build -o bin/$(NAME)-windows-amd64 .
	CGO_ENABLED=$(CGO_ENABLED) GOOS="linux" GOARCH="arm64" go build -o bin/$(NAME)-linux-aarch64 .

.PHONY : clean
clean:
	rm -f $(NAME) bin/$(NAME)-*
