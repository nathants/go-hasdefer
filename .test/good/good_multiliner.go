package good1

func main() {
	go func() {
		defer func() {}()
	}()
}
