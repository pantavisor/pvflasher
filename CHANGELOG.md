
<a name="v0.0.7"></a>
## [v0.0.7](https://github.com/pantavisor/pvflasher/compare/v0.0.6...v0.0.7)

> 2026-04-13

### Feat

* fix progress tracking and verification for compressed images
* improve device enumeration for eMMC and loop devices
* add download command to CLI
* add macOS code signing and notarization to CI release workflow
* export CLI commands for external integration


<a name="v0.0.6"></a>
## [v0.0.6](https://github.com/pantavisor/pvflasher/compare/v0.0.5...v0.0.6)

> 2026-03-31

### Feature

* add sha256sum validation of the downloaded package

### Fix

* use embed assets instead of normal assets for the pantacor logo
* make the darwin install to correctly get the binary


<a name="v0.0.5"></a>
## [v0.0.5](https://github.com/pantavisor/pvflasher/compare/v0.0.4...v0.0.5)

> 2026-03-26

### Chore

* update changelog for v0.0.5

### Feature

* update install script with better error handling and color
* add better styles to main view
* add test
* add better support to macosx
* add branding
* add system install and uninstall, and uninstall-local

### Fix

* align the footer
* remote masos sdk links
* documentation about the settings


<a name="v0.0.4"></a>
## [v0.0.4](https://github.com/pantavisor/pvflasher/compare/v0.0.3...v0.0.4)

> 2026-02-09

### Feature

* correct syncing and ejecting faces progress
* add cli install command that will download and install using images from release.json


<a name="v0.0.3"></a>
## [v0.0.3](https://github.com/pantavisor/pvflasher/compare/v0.0.2...v0.0.3)

> 2026-02-03

### Feature

* install uses the deb,rpm,and pkg when is possible

### Fix

* darwin package uses go build + fyne package --executable, ad-hoc codesign and zip -X
* correct the build mechanism for darwin
* solve shellcheck the install.sh
* install.sh script on macosx should install into ~/.pvflasher/bin/pvflasher
* install.sh tag selection for current


<a name="v0.0.2"></a>
## [v0.0.2](https://github.com/pantavisor/pvflasher/compare/v0.0.1...v0.0.2)

> 2026-02-03

### Feature

* add chglog
* save theme color selection
* build system with nfpm

### Fix

* upload-artifact glob for nfpm packages


<a name="v0.0.1"></a>
## v0.0.1

> 2026-02-02

### Feature

* add windows builds to the ci

### Fix

* all runner smoke test
* cards max size whern imasge selected

### Refactor

* split in 3 runners the builds of linux, windows, and osx

