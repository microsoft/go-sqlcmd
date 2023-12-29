package extract

import (
	"fmt"
	"testing"
)

func TestTzOutput(t *testing.T) {
	stdout := `Path = Readme_2010.txt
Size = 1157
Packed Size = 680
Modified = 2018-09-11 11:45:55.2543593
Attributes = A
CRC = B243D895
Encrypted = -
Method = LZMA2:27
Block = 0

Path = StackOverflow2010.mdf
Size = 8980398080
Packed Size = 1130813973
Modified = 2018-09-11 11:30:55.3142494
Attributes = A
CRC = 8D688B2A
Encrypted = -
Method = LZMA2:27
Block = 1

Path = StackOverflow2010_log.ldf
Size = 268312576
Packed Size = 37193161
Modified = 2018-09-11 11:30:55.3152489
Attributes = A
CRC = BCA9F91F
Encrypted = -
Method = LZMA2:27
Block = 2`

	paths := extractPaths(stdout)

	fmt.Println(paths)
}
