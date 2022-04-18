package workload

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	fuzz "github.com/google/gofuzz"
)

func TestRandomUniqueIds(t *testing.T) {
	fuzzer := fuzz.New()

	var arr []string
	for i := 0; i < 5000; i++ {
		var str int
		fuzzer.Fuzz(&str)
		arr = append(arr, strconv.Itoa(str))
	}

	fmt.Println(`""` + strings.Join(arr, `","`) + `""`)
}
