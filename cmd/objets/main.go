package main

import (
	"fmt"
	"os"

	"github.com/tsileo/objets"
)

func main() {
	// FIXME(tsileo): check if the config file and error
	if len(os.Args) < 2 {
		fmt.Printf("objets [path to conf.yaml]\n")
		return
	}
	obj, err := objets.New(os.Args[1])
	if err != nil {
		panic(err)
	}
	if err := objets.NewServer(obj).Serve(); err != nil {
		panic(err)
	}
}
