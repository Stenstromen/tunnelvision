BINARY_NAME=Tunnelvision

APP_BUNDLE=./$(BINARY_NAME).app
ENTRYPOINT_SCRIPT=entrypoint.sh

all: clean create_bundle create_plist compile_binaries create_entrypoint

clean:
	rm -rf $(APP_BUNDLE)

create_bundle:
	mkdir -p $(APP_BUNDLE)/Contents/MacOS

create_plist:
	echo '<?xml version="1.0" encoding="UTF-8"?>\n<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n<plist version="1.0">\n<dict>\n    <key>CFBundleExecutable</key>\n    <string>$(ENTRYPOINT_SCRIPT)</string>\n    <key>CFBundleIdentifier</key>\n    <string>com.example.myapp</string>\n    <key>CFBundleName</key>\n    <string>MyApp</string>\n    <key>CFBundleVersion</key>\n    <string>1.0.0</string>\n    <key>CFBundlePackageType</key>\n    <string>APPL</string>\n    <key>LSUIElement</key>\n    <string>1</string>\n</dict>\n</plist>' > $(APP_BUNDLE)/Contents/Info.plist

compile_binaries:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $(APP_BUNDLE)/Contents/MacOS/$(BINARY_NAME)_amd64
	GOOS=darwin GOARCH=arm64 go build -o $(APP_BUNDLE)/Contents/MacOS/$(BINARY_NAME)_arm64

create_entrypoint:
	echo '#!/bin/bash\n\nif [ "$$(uname -m)" = "x86_64" ]; then\n  $$(dirname "$$0")/$(BINARY_NAME)_amd64\nelse\n  $$(dirname "$$0")/$(BINARY_NAME)_arm64\nfi' > $(APP_BUNDLE)/Contents/MacOS/$(ENTRYPOINT_SCRIPT)
	chmod +x $(APP_BUNDLE)/Contents/MacOS/$(ENTRYPOINT_SCRIPT)

# Phony targets
.PHONY: all clean create_bundle create_plist compile_binaries create_entrypoint
