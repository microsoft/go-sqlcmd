// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/microsoft/go-mssqldb/msdsn"
)

// ListLocalServers queries the SQL Browser service for available SQL Server instances
// and writes the results to the provided writer.
func ListLocalServers(w io.Writer) {
	instances, err := GetLocalServerInstances()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	for _, s := range instances {
		fmt.Fprintf(w, "  %s\n", s)
	}
}

// GetLocalServerInstances queries the SQL Browser service and returns a list of
// available SQL Server instances on the local machine.
// Returns an error for non-timeout network errors.
func GetLocalServerInstances() ([]string, error) {
	bmsg := []byte{byte(msdsn.BrowserAllInstances)}
	resp := make([]byte, 16*1024-1)
	dialer := &net.Dialer{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	conn, err := dialer.DialContext(ctx, "udp", ":1434")
	// silently ignore failures to connect, same as ODBC
	if err != nil {
		return nil, nil
	}
	defer conn.Close()
	dl, _ := ctx.Deadline()
	_ = conn.SetDeadline(dl)
	_, err = conn.Write(bmsg)
	if err != nil {
		// Only return error if it's not a timeout
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			return nil, err
		}
		return nil, nil
	}
	read, err := conn.Read(resp)
	if err != nil {
		// Only return error if it's not a timeout
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			return nil, err
		}
		return nil, nil
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
	return instances, nil
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
					// Only add if InstanceName key exists and is non-empty
					if instName, ok := instanceDict["InstanceName"]; ok && instName != "" {
						results[strings.ToUpper(instName)] = instanceDict
					}
					instanceDict = map[string]string{}
					continue
				}
				gotName = true
			}
		}
	}
	return results
}
