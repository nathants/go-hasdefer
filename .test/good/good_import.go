package goodimport

type Data struct {
}

func (d *Data) Foobar2() { defer func() {}() }

func (d *Data) Foobar4() {
	defer func() {}()
}

func Foobar() {
	defer func() {

	}()
}
