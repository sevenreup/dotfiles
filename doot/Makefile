BIN     := doot
SRC     := ./src/cmd/doot
INSTALL := $(HOME)/.local/bin

.PHONY: build install clean sync

build:
	go build -o $(BIN) $(SRC)

install: build
	mkdir -p $(INSTALL)
	cp $(BIN) $(INSTALL)/$(BIN)
	@echo "Installed to $(INSTALL)/$(BIN)"

clean:
	rm -f $(BIN)

sync:
	go get github.com/emersion/go-webdav@master
	go mod tidy
