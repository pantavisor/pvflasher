
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

