package sqlcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsciiFormatter(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	if s.db == nil {
		t.Skip("No database connection available")
	}
	defer buf.Close()
	
	// Set format to ascii
	s.vars.Set(SQLCMDFORMAT, "ascii")
	s.Format = NewSQLCmdDefaultFormatter(s.vars, false, ControlIgnore)
	
	err := runSqlCmd(t, s, []string{"select 1 as id, 'test' as name", "GO"})
	assert.NoError(t, err, "runSqlCmd returned error")
	
	expected := `+----+------+` + SqlcmdEol +
		`| id | name |` + SqlcmdEol +
		`+----+------+` + SqlcmdEol +
		`|  1 | test |` + SqlcmdEol +
		`+----+------+` + SqlcmdEol +
		`(1 row affected)` + SqlcmdEol
		
	assert.Equal(t, expected, buf.buf.String())
}

func TestAsciiFormatterWrapping(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	if s.db == nil {
		t.Skip("No database connection available")
	}
	defer buf.Close()
	
	s.vars.Set(SQLCMDFORMAT, "ascii")
	s.vars.Set(SQLCMDCOLWIDTH, "20") // Small width to force wrapping
	s.Format = NewSQLCmdDefaultFormatter(s.vars, false, ControlIgnore)
	
	// Select 3 columns that won't fit in 20 chars
	err := runSqlCmd(t, s, []string{"select 1 as id, 'test' as name, '0123456789' as descr", "GO"})
	assert.NoError(t, err, "runSqlCmd returned error")
	
	expectedPart1 := `+----+------+` + SqlcmdEol +
		`| id | name |` + SqlcmdEol +
		`+----+------+` + SqlcmdEol +
		`|  1 | test |` + SqlcmdEol +
		`+----+------+` + SqlcmdEol
		
	expectedPart2 := `+------------+` + SqlcmdEol +
		`| descr      |` + SqlcmdEol +
		`+------------+` + SqlcmdEol +
		`| 0123456789 |` + SqlcmdEol +
		`+------------+` + SqlcmdEol +
		`(1 row affected)` + SqlcmdEol
		
	assert.Contains(t, buf.buf.String(), expectedPart1)
	assert.Contains(t, buf.buf.String(), expectedPart2)
}
