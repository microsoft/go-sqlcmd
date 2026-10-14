// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"strconv"
	"strings"
	"unsafe"

	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	localDBInstalledVersionsKey = `SOFTWARE\Microsoft\Microsoft SQL Server Local DB\Installed Versions`
	// LOCALDB_MAX_SQLCONNECTION_BUFFER_SIZE from sqluserinstance.h
	localDBMaxSQLConnectionBufferSize = 260
	// LOCALDB_TRUNCATE_ERR_MESSAGE from sqluserinstance.h
	localDBTruncateErrMessage  = 1
	localDBMaxErrorMessageSize = 1024
)

// startLocalDBInstance uses the LocalDB instance API to start the named instance,
// creating it first if it's an automatic instance, and returns its pipe name
func startLocalDBInstance(instance string) (string, error) {
	apiPath, err := localDBInstanceAPIPath()
	if err != nil {
		return "", err
	}
	dll, err := windows.LoadDLL(apiPath)
	if err != nil {
		return "", localizer.Errorf("Unable to load the LocalDB instance API %s: %v", apiPath, err)
	}
	defer func() { _ = dll.Release() }()
	start, err := dll.FindProc("LocalDBStartInstance")
	if err != nil {
		return "", err
	}
	instanceName, err := windows.UTF16PtrFromString(instance)
	if err != nil {
		return "", err
	}
	buf := make([]uint16, localDBMaxSQLConnectionBufferSize+1)
	size := uint32(len(buf))
	hr, _, _ := start.Call(
		uintptr(unsafe.Pointer(instanceName)),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)))
	if int32(hr) < 0 {
		message := localDBErrorMessage(dll, hr)
		return "", localizer.Errorf("Unable to start LocalDB instance %s: %s", instance, message)
	}
	return windows.UTF16ToString(buf), nil
}

// localDBInstanceAPIPath returns the path of SqlUserInstance.dll from the newest installed LocalDB version
func localDBInstanceAPIPath() (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, localDBInstalledVersionsKey, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return "", errLocalDBNotInstalled()
	}
	defer func() { _ = key.Close() }()
	versions, err := key.ReadSubKeyNames(-1)
	if err != nil || len(versions) == 0 {
		return "", errLocalDBNotInstalled()
	}
	newest := ""
	for _, v := range versions {
		if newest == "" || compareVersions(v, newest) > 0 {
			newest = v
		}
	}
	versionKey, err := registry.OpenKey(key, newest, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer func() { _ = versionKey.Close() }()
	apiPath, _, err := versionKey.GetStringValue("InstanceAPIPath")
	if err != nil {
		return "", errLocalDBNotInstalled()
	}
	return apiPath, nil
}

func errLocalDBNotInstalled() error {
	return localizer.Errorf("SQL Server Express LocalDB is not installed")
}

// compareVersions compares dotted numeric version strings such as "15.0" and "16.0"
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := 0; i < len(aParts) || i < len(bParts); i++ {
		var x, y int
		if i < len(aParts) {
			x, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			y, _ = strconv.Atoi(bParts[i])
		}
		if x != y {
			return x - y
		}
	}
	return 0
}

// localDBErrorMessage returns the LocalDB API's description of hr
func localDBErrorMessage(dll *windows.DLL, hr uintptr) string {
	fallback := "HRESULT 0x" + strconv.FormatUint(uint64(uint32(hr)), 16)
	format, err := dll.FindProc("LocalDBFormatMessage")
	if err != nil {
		return fallback
	}
	buf := make([]uint16, localDBMaxErrorMessageSize)
	size := uint32(len(buf))
	ret, _, _ := format.Call(
		hr,
		localDBTruncateErrMessage,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)))
	if int32(ret) < 0 {
		return fallback
	}
	return strings.TrimSpace(windows.UTF16ToString(buf))
}
