PROJECT_NAME := "gosysmon"
BUILD_FILES := $(shell find . -name '*.go' | grep -v _test.go)

all: build-client build

build-client:
	@echo "[*] start building client component"
	@cd client; npm install; npm run build

build: $(BUILD_FILES)
	@echo "[*] start building server component"
	@go build -o $(PROJECT_NAME) $(BUILD_FILES)

clean:
	@rm -f $(PROJECT_NAME)
