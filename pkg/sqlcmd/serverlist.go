// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/microsoft/go-mssqldb/msdsn"
)

// ListLocalServers queries the SQL Browser service for available SQL Server instances
// and writes the results to the provided writer.
func ListLocalServers(w io.Writer) {
	instances := GetLocalServerInstances()
	for _, s := range instances {
		fmt.Fprintln(w, "  ", s)
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
	for s := range data {
		if s == "MSSQLSERVER" {
			instances = append(instances, "(local)", data[s]["ServerName"])
		} else {
			instances = append(instances, fmt.Sprintf(`%s\%s`, data[s]["ServerName"], s))
		}
	}
	return instances
}

func parseInstances(msg []byte) msdsn.BrowserData {
	results := msdsn.BrowserData{}
	if len(msg) > 3 && msg[0] == 5 {
		out_s := string(msg[3:])
		tokens := strings.Split(out_s, ";")
		instdict := map[string]string{}
		got_name := false
		var name string
		for _, token := range tokens {
			if got_name {
				instdict[name] = token
				got_name = false
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
				got_name = true
			}
		}
	}
	return results
}
