package pkg1 //nolint
import (
	"fmt"
	"os"
)

func mulfunc(i int) (int, error) { //nolint
	return i * 2, nil
}

func errCheckFunc() { //nolint
	// формулируем ожидания: анализатор должен находить ошибку,
	// описанную в комментарии want
	mulfunc(5)           //nolint // want "expression returns unchecked error"
	res, _ := mulfunc(5) //nolint // want "assignment with unchecked error"
	fmt.Println(res)     //nolint // want "expression returns unchecked error"
	os.Exit(0)           //nolint // want "not allowed using of os.Exit()"
}
