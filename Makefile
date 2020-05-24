
CGO_ENABLED:=0

debug:
	CGO_ENABLED=$(CGO_ENABLED) go build -o proxysocket .

.PHONY : all

all: clean binary

binary:
	@test -d bin || mkdir bin
	CGO_ENABLED=$(CGO_ENABLED) GOOS="linux" go build -o bin/proxysocket-linux-amd64 .
	CGO_ENABLED=$(CGO_ENABLED) GOOS="darwin" go build -o bin/proxysocket-darwin-amd64 .
	CGO_ENABLED=$(CGO_ENABLED) GOOS="windows" go build -o bin/proxysocket-windows-amd64 .
	CGO_ENABLED=$(CGO_ENABLED) GOOS="linux" GOARCH="arm64" go build -o bin/proxysocket-linux-aarch64 .

.PHONY : clean
clean:
	rm -f bin/proxysocket-*
