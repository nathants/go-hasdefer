package good1

import (
	"github.com/nathants/go-hasdefer/test/goodimport"
)

func main() {
	Foobar3 := func() {
		defer func() {}()
	}

	Foobar4 := func() { defer func() {}() }

	go goodimport.Foobar()
	go goodimport.Foobar2()
	go Foobar3()
	go Foobar4()
}
