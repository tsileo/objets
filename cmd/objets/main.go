package main

import (
	"os"

	"github.com/tsileo/objets"
)

func main() {
	obj, err := objets.New(os.Args[1])
	if err != nil {
		panic(err)
	}
	if err := objets.NewServer(obj).Serve(); err != nil {
		panic(err)
	}
}
