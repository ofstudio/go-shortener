package pkg_not_main

import "os"

func NotMain() {
	os.Exit(0) // should pass the test
}
