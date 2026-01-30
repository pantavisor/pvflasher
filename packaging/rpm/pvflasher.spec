Name:           pvflasher
Version:        1.0.0
Release:        1%{?dist}
Summary:        Cross-platform USB Image pvflasher

License:        MIT
URL:            https://github.com/pantacor/pvflasher
Source0:        %{name}-%{version}.tar.gz

Requires:       libX11, libXcursor, libXrandr, libXinerama, libXi, libXxf86vm, mesa-libGL

%define debug_package %{nil}

%description
pvflasher is a fast and reliable USB image pvpvflasher with support for
bmap optimization, verification, and multiple image formats.

Features:
- Bmap-optimized flashing for faster writes
- Support for compressed images (gz, bz2, xz, zstd)
- Verification after flash
- Safe device ejection
- GUI and CLI interfaces

%prep
%setup -q

%build
# Binary should be pre-built

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_datadir}/applications
mkdir -p %{buildroot}%{_datadir}/icons/hicolor/256x256/apps

install -m 755 pvflasher %{buildroot}%{_bindir}/pvflasher
install -m 644 pvflasher.desktop %{buildroot}%{_datadir}/applications/
install -m 644 appicon.png %{buildroot}%{_datadir}/icons/hicolor/256x256/apps/pvflasher.png

%files
%{_bindir}/pvflasher
%{_datadir}/applications/pvflasher.desktop
%{_datadir}/icons/hicolor/256x256/apps/pvflasher.png

%changelog
* Wed Jan 22 2025 Sergio Marin <sergio.marin@pantacor.com> - 1.0.0-1
- Initial release
