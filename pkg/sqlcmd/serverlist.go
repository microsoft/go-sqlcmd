// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/microsoft/go-mssqldb/msdsn"
)

// ListLocalServers queries the SQL Browser service for available SQL Server instances
// and writes the results to the provided writer.
func ListLocalServers(w io.Writer) {
	instances := GetLocalServerInstances()
	for _, s := range instances {
		fmt.Fprintf(w, "  %s\n", s)
	}
}

// GetLocalServerInstances queries the SQL Browser service and returns a list of
// available SQL Server instances on the local machine.
func GetLocalServerInstances() []string {
	bmsg := []byte{byte(msdsn.BrowserAllInstances)}
	resp := make([]byte, 16*1024-1)
	dialer := &net.Dialer{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	conn, err := dialer.DialContext(ctx, "udp", ":1434")
	// silently ignore failures to connect, same as ODBC
	if err != nil {
		return nil
	}
	defer conn.Close()
	dl, _ := ctx.Deadline()
	_ = conn.SetDeadline(dl)
	_, err = conn.Write(bmsg)
	if err != nil {
		// Silently ignore errors, same as ODBC
		return nil
	}
	read, err := conn.Read(resp)
	if err != nil {
		// Silently ignore errors, same as ODBC
		return nil
	}

	data := parseInstances(resp[:read])
	instances := make([]string, 0, len(data))

	// Sort instance names for deterministic output
	instanceNames := make([]string, 0, len(data))
	for s := range data {
		instanceNames = append(instanceNames, s)
	}
	sort.Strings(instanceNames)

	for _, s := range instanceNames {
		serverName := data[s]["ServerName"]
		if serverName == "" {
			// Skip instances without a ServerName
			continue
		}
		if s == "MSSQLSERVER" {
			instances = append(instances, "(local)", serverName)
		} else {
			instances = append(instances, fmt.Sprintf(`%s\%s`, serverName, s))
		}
	}
	return instances
}

func parseInstances(msg []byte) msdsn.BrowserData {
	results := msdsn.BrowserData{}
	if len(msg) > 3 && msg[0] == 5 {
		outStr := string(msg[3:])
		tokens := strings.Split(outStr, ";")
		instanceDict := map[string]string{}
		gotName := false
		var name string
		for _, token := range tokens {
			if gotName {
				instanceDict[name] = token
				gotName = false
			} else {
				name = token
				if len(name) == 0 {
					if len(instanceDict) == 0 {
						break
					}
					results[strings.ToUpper(instanceDict["InstanceName"])] = instanceDict
					instanceDict = map[string]string{}
					continue
				}
				gotName = true
			}
		}
	}
	return results
}
