package fab

import (
	. "fmt"
)

func Memo(arg string) {
	Println("\033[0;32m  -- "+arg+" --\033[0m")
}
