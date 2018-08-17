GOCMD=go
UPX=upx
GOBUILD=$(GOCMD) build -ldflags="-s -w"
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=tm
BINARY_OSX=$(BINARY_NAME)_osx
BINARY_WIN=$(BINARY_NAME)_win

all: test build
build: 
	$(GOBUILD) -o $(BINARY_NAME) -v

shrink:
	$(UPX) $(BINARY_NAME)

test: 
	$(GOTEST) -v ./...

clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_OSX)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v 
	./$(BINARY_NAME)

validation:	
	./script/validate-vet
	./script/validate-lint
	./script/validate-gofmt
	./script/validate-git-marks

# TODO
# deps:

build-osx:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_OSX) -v