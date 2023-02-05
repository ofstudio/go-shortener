package main

import "os"

func main() {
	os.Exit(0) // want "os.Exit call in main function of main package"
}
