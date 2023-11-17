package main

import "github.com/julydate/acmeDeliver/cmd"

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
