.PHONY: build bundle install clean run

BINARY   := glitch-notify
APP_NAME := glitch-notify.app
APP_DIR  := $(HOME)/.local/Applications
BUNDLE   := $(APP_DIR)/$(APP_NAME)

build:
	swift build -c release

bundle: build
	mkdir -p "$(BUNDLE)/Contents/MacOS"
	mkdir -p "$(BUNDLE)/Contents/Resources"
	cp .build/release/GlitchNotify "$(BUNDLE)/Contents/MacOS/$(BINARY)"
	cp Info.plist "$(BUNDLE)/Contents/Info.plist"
	cp assets/icon.png "$(BUNDLE)/Contents/Resources/icon.png"
	@echo "bundled: $(BUNDLE)"

install: bundle
	ln -sf "$(BUNDLE)/Contents/MacOS/$(BINARY)" "$(HOME)/.local/bin/$(BINARY)"
	@echo "installed: $(HOME)/.local/bin/$(BINARY)"

run: bundle
	open "$(BUNDLE)"

clean:
	swift package clean
	rm -rf "$(BUNDLE)"
