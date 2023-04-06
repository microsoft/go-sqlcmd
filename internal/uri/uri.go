package uri

import (
	"path"
	"path/filepath"
	"strings"
)

func (u Uri) IsLocal() bool {
	if len(u.Scheme()) > 2 {
		return false
	} else {
		return true
	}
}

func (u Uri) ActualUrl() string {
	urlEndIdx := strings.LastIndex(u.uri, ".bak")
	if urlEndIdx == -1 {
		urlEndIdx = strings.LastIndex(u.uri, ".mdf")
	}
	if urlEndIdx != -1 {
		return u.uri[0:(urlEndIdx + 4)]
	}

	if urlEndIdx == -1 {
		urlEndIdx = strings.LastIndex(u.uri, ".7z")
		if urlEndIdx != -1 {
			return u.uri[0:(urlEndIdx + 3)]
		}
	}

	if urlEndIdx == -1 {
		urlEndIdx = strings.LastIndex(u.uri, ".bacpac")
		if urlEndIdx != -1 {
			return u.uri[0:(urlEndIdx + 7)]
		}
	}

	return u.uri
}

func (u Uri) Scheme() string {
	return u.url.Scheme
}

func (u Uri) FileExtension() string {
	_, f := filepath.Split(u.ActualUrl())
	return strings.TrimLeft(filepath.Ext(f), ".")
}

func (u Uri) Filename() string {
	filename := filepath.Base(u.ActualUrl())
	if filename == "" {
		panic("filename is empty")
	}
	return filename
}

// parseDbName returns the databaseName from --using arg
// It sets database name to the specified database name
// or in absence of it, it is set to the filename without
// extension.
func (u Uri) ParseDbName() string {
	if u.uri == "" {
		panic("uri is empty")
	}

	dbToken := path.Base(u.url.Path)
	if dbToken != "." && dbToken != "/" {
		lastIdx := strings.LastIndex(dbToken, ".bak")
		if lastIdx == -1 {
			lastIdx = strings.LastIndex(dbToken, ".mdf")
		}
		if lastIdx != -1 {
			//Get file name without extension
			fileName := dbToken[0:lastIdx]
			lastIdx += 5
			if lastIdx >= len(dbToken) {
				return fileName
			}
			//Return database name if it was specified
			return dbToken[lastIdx:]
		} else {
			lastIdx := strings.LastIndex(dbToken, ".bacpac")
			if lastIdx != -1 {
				//Get file name without extension
				fileName := dbToken[0:lastIdx]
				lastIdx += 8
				if lastIdx >= len(dbToken) {
					return fileName
				}
				//Return database name if it was specified
				return dbToken[lastIdx:]
			} else {
				lastIdx := strings.LastIndex(dbToken, ".7z")
				if lastIdx != -1 {
					//Get file name without extension
					fileName := dbToken[0:lastIdx]
					lastIdx += 4
					if lastIdx >= len(dbToken) {
						return fileName
					}
					//Return database name if it was specified
					return dbToken[lastIdx:]
				}
			}
		}
	}

	fileName := filepath.Base(u.uri)
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func (u Uri) GetDbNameAsIdentifier() string {
	escapedDbName := strings.ReplaceAll(u.ParseDbName(), "'", "''")
	dbName := strings.ReplaceAll(escapedDbName, "]", "]]")
	if dbName == "" {
		panic("database name is empty")
	}
	return dbName
}

func (u Uri) GetDbNameAsNonIdentifier() string {
	dbName := strings.ReplaceAll(u.ParseDbName(), "]", "]]")
	if dbName == "" {
		panic("database name is empty")
	}
	return dbName
}
