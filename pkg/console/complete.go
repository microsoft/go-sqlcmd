// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package console

import "strings"

// CompleteLine returns a set of candidate TSQL keywords to complete the current input line
func CompleteLine(line string) []string {
	idx := strings.LastIndexAny(line, " ;") + 1
	// we don't try to complete without a starting letter
	if idx == len(line) {
		return []string{}
	}
	prefix := strings.ToUpper(string(line[idx:]))
	left := 0
	right := len(keywords) - 1
	for left <= right {
		mid := (left + right) / 2
		comp := 0
		if len(keywords[mid]) >= len(prefix) {
			comp = strings.Compare(prefix, string(keywords[mid][:len(prefix)]))
		} else {
			comp = strings.Compare(prefix, keywords[mid])
		}
		if comp < 0 {
			right = mid - 1
		} else if comp > 0 {
			left = mid + 1
		} else {
			// look up and down the list from mid and return the slice of matching words
			first := mid - 1
			last := mid + 1
			for first >= 0 && strings.HasPrefix(keywords[first], prefix) {
				first--
			}
			for last < len(keywords) && strings.HasPrefix(keywords[last], prefix) {
				last++
			}
			lines := make([]string, last-first-1)
			for i, w := range keywords[first+1 : last] {
				lines[i] = mergeLine(line, w, idx)
			}
			return lines
		}
	}
	return []string{}
}

// mergeline appends keyword to line starting at index idx
// It matches the case of the current character in the line
func mergeLine(line string, keyword string, idx int) string {
	upcase := line[idx] >= 'A' && line[idx] <= 'Z'
	b := strings.Builder{}
	b.Write([]byte(line[:idx]))
	if !upcase {
		b.WriteString(strings.ToLower(keyword))
	} else {
		b.WriteString(keyword)
	}
	return b.String()
}

var keywords = []string{
	"ADD",
	"ALL",
	"ALTER",
	"AND",
	"ANY",
	"AS",
	"ASC",
	"AUTHORIZATION",
	"BACKUP",
	"BEGIN",
	"BETWEEN",
	"BREAK",
	"BROWSE",
	"BULK",
	"BY",
	"CASCADE",
	"CASE",
	"CHECK",
	"CHECKPOINT",
	"CLOSE",
	"CLUSTERED",
	"COALESCE",
	"COLLATE",
	"COLUMN",
	"COMMIT",
	"COMPUTE",
	"CONSTRAINT",
	"CONTAINS",
	"CONTAINSTABLE",
	"CONTINUE",
	"CONVERT",
	"CREATE",
	"CROSS",
	"CURRENT",
	"CURRENT_DATE",
	"CURRENT_TIME",
	"CURRENT_TIMESTAMP",
	"CURRENT_USER",
	"CURSOR",
	"DATABASE",
	"DBCC",
	"DEALLOCATE",
	"DECLARE",
	"DEFAULT",
	"DELETE",
	"DENY",
	"DESC",
	"DISTINCT",
	"DISTRIBUTED",
	"DOUBLE",
	"DROP",
	"ELSE",
	"END",
	"ERRLVL",
	"ESCAPE",
	"EXCEPT",
	"EXEC",
	"EXECUTE",
	"EXISTS",
	"EXIT",
	"EXTERNAL",
	"FETCH",
	"FILE",
	"FILLFACTOR",
	"FOR",
	"FOREIGN",
	"FREETEXT",
	"FREETEXTTABLE",
	"FROM",
	"FULL",
	"FUNCTION",
	"GOTO",
	"GRANT",
	"GROUP",
	"HAVING",
	"HOLDLOCK",
	"IDENTITY",
	"IDENTITY_INSERT",
	"IDENTITYCOL",
	"IF",
	"IN",
	"INDEX",
	"INNER",
	"INSERT",
	"INTERSECT",
	"INTO",
	"IS",
	"JOIN",
	"KEY",
	"KILL",
	"LEFT",
	"LIKE",
	"LINENO",
	"MERGE",
	"NATIONAL",
	"NOCHECK",
	"NONCLUSTERED",
	"NOT",
	"NULL",
	"NULLIF",
	"OF",
	"OFF",
	"OFFSETS",
	"ON",
	"OPEN",
	"OPENDATASOURCE",
	"OPENQUERY",
	"OPENROWSET",
	"OPENXML",
	"OPTION",
	"OR",
	"ORDER",
	"OUTER",
	"OVER",
	"PERCENT",
	"PIVOT",
	"PLAN",
	"PRIMARY",
	"PRINT",
	"PROC",
	"PROCEDURE",
	"PUBLIC",
	"RAISERROR",
	"READ",
	"READTEXT",
	"RECONFIGURE",
	"REFERENCES",
	"REPLICATION",
	"RESTORE",
	"RESTRICT",
	"RETURN",
	"REVERT",
	"REVOKE",
	"RIGHT",
	"ROLLBACK",
	"ROWCOUNT",
	"ROWGUIDCOL",
	"RULE",
	"SAVE",
	"SCHEMA",
	"SELECT",
	"SEMANTICKEYPHRASETABLE",
	"SEMANTICSIMILARITYDETAILSTABLE",
	"SEMANTICSIMILARITYTABLE",
	"SESSION_USER",
	"SET",
	"SETUSER",
	"SHUTDOWN",
	"SOME",
	"STATISTICS",
	"SYSTEM_USER",
	"TABLE",
	"TABLESAMPLE",
	"TEXTSIZE",
	"THEN",
	"TO",
	"TOP",
	"TRAN",
	"TRANSACTION",
	"TRIGGER",
	"TRUNCATE",
	"TRY_CONVERT",
	"TSEQUAL",
	"UNION",
	"UNIQUE",
	"UNPIVOT",
	"UPDATE",
	"UPDATETEXT",
	"USE",
	"USER",
	"VALUES",
	"VARYING",
	"VIEW",
	"WAITFOR",
	"WHEN",
	"WHERE",
	"WHERECURRENT",
	"WHILE",
	"WITH",
	"WRITETEXT",
}
