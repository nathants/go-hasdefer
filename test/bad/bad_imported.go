package bad1

import (
	"github.com/nathants/go-hasdefer/test/badimport"
)

func main() {
	Foobar3 := func() {

	}

	Foobar4 := func() {}

	go badimport.Foobar()
	go badimport.Foobar2()
	go Foobar3()
	go Foobar4()
}
