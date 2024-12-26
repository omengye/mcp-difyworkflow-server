.PHONY: build

build:
ifndef GOOS
ifndef GOARCH
	@echo "Building for all platforms..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o build/mcp-difyworkflow-server_win_AMD64 .
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/mcp-difyworkflow-server_macOS_amd64 .
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o build/mcp-difyworkflow-server_macOS_ARM64 .
else
	@echo "Building for specific platform: GOOS=$(GOOS), GOARCH=$(GOARCH)"
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o build/mcp-server-difyworkflow_$(GOOS)_$(GOARCH) .
endif
endif