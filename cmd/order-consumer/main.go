package main

import "log"

func main() {
	err := run()
	if err != nil {
		log.Fatalf("error from run: %v", err)
	}
}
