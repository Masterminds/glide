%define debug_package %{nil}

%if %{?_glide_version:1}0
%define glide_version %{_glide_version}
%else
%define glide_version 0.11.1
%endif

Name:		glide
Version:	%(echo %{glide_version}|tr - .)
Release:	1%{?dist}
Summary:	Vendor Package Management for Golang

Group:		devel
License:	Development/Libraries
URL:		https://github.com/Masterminds/glide
Source0:	glide-%{glide_version}-src.tar.gz
BuildRoot:	%(mktemp -ud %{_tmppath}/%{name}-%{version}-%{release}-XXXXXX)

BuildRequires:	golang
BuildRequires:	rsync
#Requires:	

%description
Glide is a tool for managing the vendor directory within a Go package. This
feature, first introduced in Go 1.5, allows each package to have a vendor
directory containing dependent packages for the project. These vendor packages
can be installed by a tool (e.g. glide), similar to go get or they can be
vendored and distributed with the package.


%prep
%setup -q -n %{name}-%{glide_version}

%build
mkdir -p gopath/src/github.com/Masterminds/glide
export GOPATH=${PWD}/gopath
export PATH=${GOPATH}:${PATH}
rsync -az --exclude=gopath/ ./ gopath/src/github.com/Masterminds/glide/
cd gopath/src/github.com/Masterminds/glide
make %{?_smp_mflags} VERSION=${RPM_PACKAGE_VERSION}


%install
export GOPATH=${PWD}/gopath
export PATH=${GOPATH}:${PATH}
rm -rf %{buildroot}
cd gopath/src/github.com/Masterminds/glide
make install DESTDIR=%{buildroot} PREFIX=/usr VERSION=${RPM_PACKAGE_VERSION}


%clean
rm -rf %{buildroot}

%check
export GOPATH=${PWD}/gopath
export PATH=${GOPATH}:${PATH}
cd gopath/src/github.com/Masterminds/glide
make test

%files
%defattr(-,root,root,-)
/usr/bin/glide
%doc /usr/share/doc/glide/LICENSE
%doc /usr/share/doc/glide/README.md

%changelog
* Thu Aug 4 2016 Andrii Senkovych <jolly_roger@itblog.org.ua> - 0.11.1-1
- Initial commit
