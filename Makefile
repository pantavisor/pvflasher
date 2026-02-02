BINARY_NAME=pvflasher
VERSION ?= $(shell git describe --tags --always --dirty="-dev" 2>/dev/null || echo "1.0.0")
LDFLAGS := -X=pvflasher/internal/version.Version=$(VERSION)

LINUX_AMD64_LIBS = /usr/lib /usr/lib64 /usr/lib/x86_64-linux-gnu
LINUX_ARM64_LIBS = /usr/lib /usr/lib64 /usr/lib/aarch64-linux-gnu
LINUX_ARM_LIBS = /usr/lib /usr/lib64 /usr/lib/arm-linux-gnueabihf
LINUX_AMD64_LDFLAGS = $(foreach dir,$(wildcard $(LINUX_AMD64_LIBS)),-L$(dir))
LINUX_ARM64_LDFLAGS = $(foreach dir,$(wildcard $(LINUX_ARM64_LIBS)),-L$(dir))
LINUX_ARM_LDFLAGS = $(foreach dir,$(wildcard $(LINUX_ARM_LIBS)),-L$(dir))

EXE_EXT_windows = .exe
EXE_EXT_linux =
EXE_EXT_darwin =

TAGS_linux = -tags flatpak
TAGS_windows =
TAGS_darwin =

LDFLAGS_windows = -H=windowsgui

ZIG_CC_FLAGS_windows = -Wdeprecated-non-prototype -Wl,--subsystem,windows
ZIG_CC_FLAGS_linux-amd64 = -isystem /usr/include $(LINUX_AMD64_LDFLAGS) # Native build
ZIG_CC_FLAGS_linux-arm64 = -isystem /usr/include $(LINUX_ARM64_LDFLAGS)
ZIG_CC_FLAGS_linux-arm = -isystem /usr/include $(LINUX_ARM_LDFLAGS)

ZIG_TARGET_linux-amd64 = x86_64-linux-gnu
ZIG_TARGET_linux-arm64 = aarch64-linux-gnu
ZIG_TARGET_linux-arm = arm-linux-gnueabihf
ZIG_TARGET_windows-amd64 = x86_64-windows-gnu
ZIG_TARGET_windows-arm64 = aarch64-windows-gnu

.PHONY: all build clean test run package-appimage package-deb package-rpm install-local help debian-deps

all: build

build:
	@echo "Building native pvflasher $(VERSION)..."
	@mkdir -p bin
	go build -tags flatpak -o bin/$(BINARY_NAME) -ldflags "$(LDFLAGS)"


release: release-linux-amd64 release-linux-arm64 release-windows-amd64 release-windows-arm64 release-darwin-amd64 release-darwin-arm64

# CI target: builds only the core packages using fyne-cross, plus AppImage on host
release-ci: package-linux-amd64 package-linux-arm64 package-windows-amd64 package-windows-arm64 package-appimage-host-amd64 package-appimage-host-arm64
	@echo "Gathering CI release artifacts..."
	@mkdir -p release/linux release/windows
	@if [ -f fyne-cross/dist/linux-amd64/$(BINARY_NAME).tar.xz ]; then \
		cp fyne-cross/dist/linux-amd64/$(BINARY_NAME).tar.xz release/linux/$(BINARY_NAME)-linux-amd64.tar.xz; \
	fi
	@if [ -f fyne-cross/dist/linux-arm64/$(BINARY_NAME).tar.xz ]; then \
		cp fyne-cross/dist/linux-arm64/$(BINARY_NAME).tar.xz release/linux/$(BINARY_NAME)-linux-arm64.tar.xz; \
	fi
	@if [ -f fyne-cross/dist/windows-amd64/$(BINARY_NAME).zip ]; then \
		cp fyne-cross/dist/windows-amd64/$(BINARY_NAME).zip release/windows/$(BINARY_NAME)-windows-amd64.zip; \
	fi
	@if [ -f fyne-cross/dist/windows-arm64/$(BINARY_NAME).zip ]; then \
		cp fyne-cross/dist/windows-arm64/$(BINARY_NAME).zip release/windows/$(BINARY_NAME)-windows-arm64.zip; \
	fi
	@echo "CI Artifacts available in release/"

# macOS CI target: builds macOS packages natively on macOS runner
release-ci-darwin: package-darwin-native-amd64 package-darwin-native-arm64
	@echo "macOS CI Artifacts available in release/darwin/"

# Native macOS packaging (no Docker required, for use on macOS hosts)
package-darwin-native-%:
	@echo "Building native macOS app for $*..."
	@mkdir -p release/darwin
	GOARCH=$* CGO_ENABLED=1 \
	fyne package -os darwin \
		-name $(BINARY_NAME) \
		-icon Icon.png \
		-appID com.pantacor.pvflasher
	@mkdir -p release/darwin/$(BINARY_NAME)-darwin-$*
	@mv $(BINARY_NAME).app release/darwin/$(BINARY_NAME)-darwin-$*/
	@cd release/darwin && zip -r $(BINARY_NAME)-darwin-$*.zip $(BINARY_NAME)-darwin-$*
	@rm -rf release/darwin/$(BINARY_NAME)-darwin-$*
	@echo "Created release/darwin/$(BINARY_NAME)-darwin-$*.zip"

# Host-based AppImage packaging (requires linuxdeploy and appimagetool on host)
package-appimage-host-%: package-linux-%
	@echo "Building AppImage for $* on host..."
	@mkdir -p release/linux
	@chmod +x scripts/*.sh
	bash ./scripts/build-appimage-host.sh $(VERSION) $*
	@rm -rf release/linux-$*
	@rm -rf packaging/appimage/PvFlasher.AppDir

# Specialized target for Linux AMD64 to include all packages
release-linux-amd64: package-linux-amd64 package-appimage-amd64 package-deb-amd64 package-rpm-amd64
	@echo "Gathering release artifacts for Linux AMD64..."
	@mkdir -p release/linux
	@if [ -f fyne-cross/dist/linux-amd64/$(BINARY_NAME).tar.xz ]; then \
		cp fyne-cross/dist/linux-amd64/$(BINARY_NAME).tar.xz release/linux/$(BINARY_NAME)-linux-amd64.tar.xz; \
	fi
	@echo "Artifacts available in release/linux/"

# Specialized target for Linux ARM64 to include all packages
release-linux-arm64: package-linux-arm64 package-appimage-arm64 package-deb-arm64 package-rpm-arm64
	@echo "Gathering release artifacts for Linux ARM64..."
	@mkdir -p release/linux
	@if [ -f fyne-cross/dist/linux-arm64/$(BINARY_NAME).tar.xz ]; then \
		cp fyne-cross/dist/linux-arm64/$(BINARY_NAME).tar.xz release/linux/$(BINARY_NAME)-linux-arm64.tar.xz; \
	fi
	@echo "Artifacts available in release/linux/"

release-linux-%: package-linux-%
	@echo "Gathering release artifacts for Linux $*..."
	@mkdir -p release/linux
	@if [ -f fyne-cross/dist/linux-$*/$(BINARY_NAME).tar.xz ]; then \
		cp fyne-cross/dist/linux-$*/$(BINARY_NAME).tar.xz release/linux/$(BINARY_NAME)-linux-$*.tar.xz; \
	fi

release-windows-%: package-windows-%
	@echo "Gathering release artifacts for Windows $*..."
	@mkdir -p release/windows
	@if [ -f fyne-cross/dist/windows-$*/$(BINARY_NAME).zip ]; then \
		cp fyne-cross/dist/windows-$*/$(BINARY_NAME).zip release/windows/$(BINARY_NAME)-windows-$*.zip; \
	fi

release-darwin-%: package-darwin-%
	@echo "Gathering release artifacts for macOS $*..."
	@mkdir -p release/darwin
	@if [ -f fyne-cross/dist/darwin-$*/$(BINARY_NAME).zip ]; then \
		cp fyne-cross/dist/darwin-$*/$(BINARY_NAME).zip release/darwin/$(BINARY_NAME)-darwin-$*.zip; \
	fi

# Generic build rule for other targets (requires sysroot if cross-compiling)
build-%:
	$(eval GOOS := $(word 1,$(subst -, ,$*)))
	$(eval GOARCH := $(word 2,$(subst -, ,$*)))
	@echo "Building for $(GOOS) ($(GOARCH)) using Zig..."
	@mkdir -p bin
	CGO_ENABLED=1 \
	CC="zig cc -target $(ZIG_TARGET_$*)" \
	CXX="zig c++ -target $(ZIG_TARGET_$*)" \
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go build $(TAGS_$(GOOS)) -o bin/$(BINARY_NAME)-$*$(EXE_EXT_$(GOOS)) -ldflags "$(LDFLAGS)"

run:
	go run -tags flatpak .

test:
	go test -v ./internal/... ./cli/...

clean:
	rm -rf bin/ build/ release/ fyne-cross dist
	rm -rf packaging/appimage/PvFlasher.AppDir
	rm -rf packaging/deb/pvflasher-*/
	rm -rf packaging/rpm/BUILD/ packaging/rpm/RPMS/ packaging/rpm/SOURCES/ packaging/rpm/SPECS/ packaging/rpm/SRPMS/


# Cross-platform packaging with fyne-cross
package-%: go.mod $(wildcard *.go) $(wildcard gui/*.go)
	$(eval GOOS := $(word 1,$(subst -, ,$*)))
	$(eval GOARCH := $(word 2,$(subst -, ,$*)))
	@echo "Building Fyne app for $(GOOS) ($(GOARCH))"
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 \
	fyne-cross $(GOOS) -arch=$(GOARCH) \
		-name $(BINARY_NAME)$(EXE_EXT_$(GOOS)) \
		-icon Icon.png \
		-app-id com.pantacor.pvflasher \
		-dir . \
		$(TAGS_$(GOOS)) \
		-ldflags="$(LDFLAGS)"
	@mkdir -p release/$(GOOS)-$(GOARCH)
	@cp fyne-cross/bin/$(GOOS)-$(GOARCH)/$(BINARY_NAME)$(EXE_EXT_$(GOOS)) release/$(GOOS)-$(GOARCH)/$(BINARY_NAME)$(EXE_EXT_$(GOOS))

package-appimage-amd64: package-linux-amd64
	@echo "Building AppImage for amd64 (packaging binary from fyne-cross)..."
	@mkdir -p release/linux
	@chmod +x packaging/appimage/*.sh
	cd packaging/appimage && bash ./build-appimage.sh $(VERSION) amd64
	@rm -rf release/linux-amd64
	@rm -rf packaging/appimage/PvFlasher.AppDir

package-appimage-arm64: package-linux-arm64
	@echo "Building AppImage for arm64 (packaging binary from fyne-cross)..."
	@mkdir -p release/linux
	@chmod +x packaging/appimage/*.sh
	cd packaging/appimage && bash ./build-appimage.sh $(VERSION) arm64
	@rm -rf release/linux-arm64
	@rm -rf packaging/appimage/PvFlasher.AppDir

package-deb-%: package-linux-%
	@echo "Building Debian package for $*..."
	@mkdir -p release/linux
	@chmod +x packaging/deb/*.sh
	cd packaging/deb && bash ./build-deb.sh $(VERSION) $*
	@rm -rf release/linux-$*
	@rm -rf packaging/deb/pvflasher-*

package-rpm-%: package-linux-%
	@echo "Building RPM package for $*..."
	@mkdir -p release/linux
	@chmod +x packaging/rpm/*.sh
	cd packaging/rpm && bash ./build-rpm.sh $(VERSION) $*
	@rm -rf release/linux-$*
	@rm -rf packaging/rpm/pvflasher-*/

package: package-appimage-amd64 package-deb-amd64 package-rpm-amd64

install-local: build
	@echo "Installing pvflasher locally to ~/.local..."
	@mkdir -p ~/.local/bin
	@mkdir -p ~/.local/share/applications
	@mkdir -p ~/.local/share/icons/hicolor/256x256/apps
	@install -m 755 bin/pvflasher ~/.local/bin/pvflasher
	@sed "s|\$$HOME|$(HOME)|g" packaging/pvflasher-local.desktop > ~/.local/share/applications/pvflasher.desktop
	@chmod 644 ~/.local/share/applications/pvflasher.desktop
	@install -m 644 Icon.png ~/.local/share/icons/hicolor/256x256/apps/pvflasher.png
	@echo "Installation complete!"
	@echo "Make sure ~/.local/bin is in your PATH to run: pvflasher"

help:
	@echo "pvflasher Build System (Fyne GUI)"
	@echo "=================================="
	@echo "  make build              - Build native binary"
	@echo "  make build-linux-amd64  - Build for Linux x64 using Zig"
	@echo "  make build-linux-arm64  - Build for Linux ARM64 using Zig"
	@echo "  make install-local      - Build and install locally to ~/.local"
	@echo "  make package-appimage   - Create Linux AppImage using Docker"
	@echo "  make package-deb        - Create Debian package"
	@echo "  make package-rpm        - Create RPM package"
	@echo "  make package-linux-amd64  - Create Fyne cross-platform app for Linux x64"
	@echo "  make package-windows-amd64 - Create Fyne app for Windows"
	@echo "  make package-darwin-amd64 - Create Fyne app for macOS Intel"
	@echo "  make package-darwin-arm64 - Create Fyne app for macOS Apple Silicon"
	@echo "  make run                - Development mode (direct run)"
	@echo "  make test               - Run tests"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make debian-deps        - Install Debian/Ubuntu dependencies"
