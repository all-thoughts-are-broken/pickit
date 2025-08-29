TARGET := pickit
BUILD_DIR := bin
SOURCE := ./main.go
UPX_LEVEL := --best  # UPX 压缩级别: --best (最高压缩) 或 --ultra-brute (极限压缩，但更慢)

.PHONY: all
all: windows linux compress

.PHONY: windows
windows:
	@echo "Building Windows executable..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(TARGET)-windows.exe $(SOURCE)

.PHONY: linux
linux:
	@echo "Building Linux executable..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(TARGET)-linux $(SOURCE)
	@chmod +x $(BUILD_DIR)/$(TARGET)-linux  # 确保 Linux 二进制文件有执行权限

.PHONY: compress
compress:
	@echo "Compressing binaries with UPX..."
	@if command -v upx >/dev/null 2>&1; then \
		echo "Files in build directory:"; \
		ls -la $(BUILD_DIR)/; \
		for bin in $(BUILD_DIR)/*; do \
			if [ -f "$$bin" ]; then \
				echo "Compressing $$bin..."; \
				upx $(UPX_LEVEL) "$$bin" 2>&1 | grep -v "NotCompressible" || echo "Note: $$bin may already be compressed or not compressible"; \
			fi; \
		done; \
		echo "Compression complete. Final file sizes:"; \
		ls -la $(BUILD_DIR)/; \
	else \
		echo "UPX not found. Please install UPX to compress binaries."; \
		echo "On Ubuntu/Debian: sudo apt install upx"; \
		echo "On macOS: brew install upx"; \
		exit 1; \
	fi

.PHONY: compress-only
compress-only:
	@echo "Compressing existing binaries with UPX..."
	@if command -v upx >/dev/null 2>&1; then \
		for bin in $(BUILD_DIR)/*; do \
			if [ -f "$$bin" ]; then \
				echo "Compressing $$bin..."; \
				upx $(UPX_LEVEL) "$$bin" 2>&1 | grep -v "NotCompressible" || echo "Note: $$bin may already be compressed or not compressible"; \
			fi; \
		done; \
	else \
		echo "UPX not found."; \
		exit 1; \
	fi

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)/*

.PHONY: help
help:
	@echo "Usage:"
	@echo "  make all           # Build for both Windows and Linux and compress (default)"
	@echo "  make windows       # Build only for Windows"
	@echo "  make linux         # Build only for Linux"
	@echo "  make compress      # Compress existing binaries with UPX"
	@echo "  make compress-only # Only compress without building"
	@echo "  make clean         # Clean build artifacts"
	@echo "  make help          # Show this help message"

.DEFAULT_GOAL := all