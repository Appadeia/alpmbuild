Name:    hello
EVR:     5:2.10-0
Summary: Hello World

License: GPL
URL:     https://gnu.org

Source0: https://ftp.gnu.org/gnu/hello/%{name}-%{version}.tar.gz with md5 6cd0ffea3884a4e79330338dcc2987d6
#!alpmbuild ReasonFor cyanogen: This is a really cool package
Recommends: cyanogen
%if 0
Requires: invalid
%elif 1 == 3
Requires: still-invalid
%else
Requires: base pingas
%endif
BuildRequires: gcc
BuildRequires: make
BuildRequires: gettext

%install
mkdir -p %{?buildroot}/%{_bindir}
mkdir -p %{?buildroot}/%{_datadir}/info
mkdir -p %{?buildroot}/%{_datadir}/man/man1
mkdir -p %{?buildroot}/%{_datadir}/locale/en_US/LC_MESSAGES/
touch %{?buildroot}/.dotfile
touch %{?buildroot}/%{_bindir}/hello
touch %{?buildroot}/%{_datadir}/info/dir
touch %{?buildroot}/%{_datadir}/info/hello.info
echo %{?buildroot} > %{?buildroot}/%{_datadir}/man/man1/hello.1
touch %{?buildroot}/%{_datadir}/locale/en_US/LC_MESSAGES/hello.mo
touch "%{?buildroot}/bad
filename"
chown -R root:root %{?buildroot}

%files
%{_bindir}/hello
%{_datadir}/info/dir
%{_datadir}/info/hello.info
%{_datadir}/man/man1/hello.1

%package translationfiles
Summary: Translation files for %{name}

%files translationfiles
%{_datadir}/locale/*/LC_MESSAGES/hello.mo