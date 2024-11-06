Name:           libzip
Version:        1.5.2
Release:        1
Summary:        libzip
License:        GPL
URL:            http://www.wi-stor.com/

BuildRequires:  systemd-units
Requires:       systemd
Source0: libzip-1.5.2.tar.gz

%description
wi-stor web manager.

%prep
%setup -q

%build





%install
cp -r  %{_builddir}/libzip-1.5.2/* %{buildroot}/libzip-1.5.2
install -d -m 755 %{_libdir}
install -d -m 755 %{_includedir}
install -d -m 755 %{_bindir}
install -d -m 755 %{_datadir}
install -d -m 755 %{buildroot}/libzip-1.5.2/build
cmake  %{buildroot}/libzip-1.5.2/build  -DCMAKE_INSTALL_PREFIX=%{buildroot}/usr %{buildroot}/libzip-1.5.2/
make %{buildroot}/libzip-1.5.2/build
make install %{buildroot}


%clean


%files
%defattr(-, root, root)
%{_libdir}
%{_includedir}
%{_bindir}
%{_datadir}


%doc

%changelog

