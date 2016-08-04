package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/dustin/go-jsonpointer"
)

func main() {
	d, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading json from stdin: %v", err)
	}
	if len(os.Args) == 1 {
		l, err := jsonpointer.ListPointers(d)
		if err != nil {
			log.Fatalf("Error listing pointers: %v", err)
		}
		for _, p := range l {
			fmt.Println(p)
		}
	}
}
