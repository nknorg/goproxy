BUILD=go build -ldflags "-s -w"
BUILD_DIR=build
ifdef GOARM
BIN_DIR=$(GOOS)-$(GOARCH)v$(GOARM)
else
BIN_DIR=$(GOOS)-$(GOARCH)
endif

.PHONY: build
build:
	mkdir -p $(BUILD_DIR)/$(BIN_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) $(BUILD) -o $(BUILD_DIR)/$(BIN_DIR)/goproxy$(EXT) .
	${MAKE} zip

.PHONY: tar
tar:
	cd $(BUILD_DIR) && rm -f $(BIN_DIR).tar.gz && tar --exclude ".DS_Store" --exclude "__MACOSX" -czvf $(BIN_DIR).tar.gz $(BIN_DIR)

.PHONY: zip
zip:
	cd $(BUILD_DIR) && rm -f $(BIN_DIR).zip && zip --exclude "*.DS_Store*" --exclude "*__MACOSX*" -r $(BIN_DIR).zip $(BIN_DIR)

.PHONY: all
all:
	${MAKE} build GOOS=darwin GOARCH=amd64
	${MAKE} build GOOS=linux GOARCH=arm GOARM=5
	${MAKE} build GOOS=linux GOARCH=arm GOARM=6
	${MAKE} build GOOS=linux GOARCH=arm GOARM=7
	${MAKE} build GOOS=linux GOARCH=arm64
	${MAKE} build GOOS=linux GOARCH=amd64
	${MAKE} build GOOS=windows GOARCH=amd64 EXT=.exe
