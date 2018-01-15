package libs

import (
	"testing"
)

func TestCSV(t *testing.T) {

	valid := map[string]interface{}{
		`(. (encoding/csv.NewReader (strings.NewReader "1,2,3,4")) ReadAll)`: [][]string{[]string{"1", "2", "3", "4"}},
	}

	checkSExprs(t, valid)
}
