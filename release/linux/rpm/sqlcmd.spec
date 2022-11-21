# RPM spec file for sqlcmd
# Definition of macros used - https://fedoraproject.org/wiki/Packaging:RPMMacros?rd=Packaging/RPMMacros

# .el7.centos -> .el7
%if 0%{?rhel}
  %define dist .el%{?rhel}
%endif

%define name           sqlcmd
%define release        1%{?dist}
%define version        %{getenv:CLI_VERSION}
%define repo_path      %{getenv:REPO_ROOT_DIR}
%define cli_lib_dir    %{_libdir}/sqlcmd

%undefine _missing_build_ids_terminate_build
%global _missing_build_ids_terminate_build 0

Summary:        MSSQL SQLCMD CLI Tools
License:        https://github.com/microsoft/go-sqlcmd/blob/main/LICENSE
Name:           %{name}
Version:        %{version}
Release:        %{release}
Url:            https://github.com/microsoft/go-sqlcmd
BuildArch:      x86_64

%description
SQLCMD CLI, a multi-platform command line experience for Microsoft SQL Server and Azure SQL.

%prep
%install

# Create executable
mkdir -p %{buildroot}%{_bindir}
cp %{repo_path}/sqlcmd %{buildroot}%{_bindir}

%files
%attr(0755,root,root) %{_bindir}/sqlcmd
