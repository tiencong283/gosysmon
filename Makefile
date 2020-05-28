PROJECT_NAME := "gosysmon"
BUILD_FILES := $(shell find . -name '*.go' | grep -v _test.go)

build: $(BUILD_FILES)
	@go build -o $(PROJECT_NAME) $(BUILD_FILES)

clean:
	@rm -f $(PROJECT_NAME)