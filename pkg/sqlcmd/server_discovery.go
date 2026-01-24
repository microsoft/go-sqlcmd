// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"time"

	"github.com/microsoft/go-mssqldb/msdsn"
)

// ServerInstance represents a discovered SQL Server instance
type ServerInstance struct {
	ServerName   string
	InstanceName string
	IsClustered  string
	Version      string
	Port         string
}

// ListServers discovers SQL Server instances on the network using the SQL Server Browser service.
// It sends a UDP broadcast to port 1434 and parses the response.
// Returns a slice of ServerInstance and any error encountered.
func ListServers(timeout time.Duration) ([]ServerInstance, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	bmsg := []byte{byte(msdsn.BrowserAllInstances)}
	resp := make([]byte, 16*1024-1)

	dialer := &net.Dialer{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "udp", ":1434")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	dl, _ := ctx.Deadline()
	_ = conn.SetDeadline(dl)

	_, err = conn.Write(bmsg)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return []ServerInstance{}, nil
		}
		return nil, err
	}

	read, err := conn.Read(resp)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return []ServerInstance{}, nil
		}
		return nil, err
	}

	data := parseInstanceData(resp[:read])
	instances := make([]ServerInstance, 0, len(data))

	for instName, props := range data {
		inst := ServerInstance{
			ServerName:   props["ServerName"],
			InstanceName: instName,
			IsClustered:  props["IsClustered"],
			Version:      props["Version"],
			Port:         props["tcp"],
		}
		instances = append(instances, inst)
	}

	return instances, nil
}

// parseInstanceData parses the SQL Server Browser response into a map of instance data
func parseInstanceData(msg []byte) msdsn.BrowserData {
	results := msdsn.BrowserData{}
	if len(msg) > 3 && msg[0] == 5 {
		outS := string(msg[3:])
		tokens := strings.Split(outS, ";")
		instdict := map[string]string{}
		gotName := false
		var name string
		for _, token := range tokens {
			if gotName {
				instdict[name] = token
				gotName = false
			} else {
				name = token
				if len(name) == 0 {
					if len(instdict) == 0 {
						break
					}
					results[strings.ToUpper(instdict["InstanceName"])] = instdict
					instdict = map[string]string{}
					continue
				}
				gotName = true
			}
		}
	}
	return results
}

// FormatServerList formats the list of server instances for display
func FormatServerList(instances []ServerInstance) []string {
	result := make([]string, 0, len(instances)*2)
	for _, inst := range instances {
		if inst.InstanceName == "MSSQLSERVER" {
			// Default instance - show both (local) and server name (same as ODBC sqlcmd)
			result = append(result, "(local)", inst.ServerName)
		} else {
			// Named instance
			result = append(result, inst.ServerName+"\\"+inst.InstanceName)
		}
	}
	return result
}
