package pkg_not_main

import "os"

// NotMain - вызов os.Exit в этой функции не должен вызывать ошибку
func NotMain() {
	os.Exit(0) // should pass the test
}
