BINARY_NAME=Tunnelvision

APP_BUNDLE=./$(BINARY_NAME).app

all: clean create_bundle create_plist build

clean:
	rm -rf $(APP_BUNDLE)

create_bundle:
	mkdir -p $(APP_BUNDLE)/Contents/MacOS

create_plist:
	echo '<?xml version="1.0" encoding="UTF-8"?>\n<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n<plist version="1.0">\n<dict>\n    <key>CFBundleExecutable</key>\n    <string>$(BINARY_NAME)</string>\n    <key>CFBundleIdentifier</key>\n    <string>com.example.myapp</string>\n    <key>CFBundleName</key>\n    <string>MyApp</string>\n    <key>CFBundleVersion</key>\n    <string>1.0.0</string>\n    <key>CFBundlePackageType</key>\n    <string>APPL</string>\n    <key>LSUIElement</key>\n    <string>1</string>\n</dict>\n</plist>' > $(APP_BUNDLE)/Contents/Info.plist

build:
	go build -o $(APP_BUNDLE)/Contents/MacOS/$(BINARY_NAME)

# Phony targets
.PHONY: all clean create_bundle create_plist build
