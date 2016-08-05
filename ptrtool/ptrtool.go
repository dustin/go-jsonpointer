package main

import (
	"bytes"
	"encoding/json"
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
	} else {
		m, err := jsonpointer.FindMany(d, os.Args[1:])
		if err != nil {
			log.Fatalf("Error finding pointers: %v", err)
		}
		for k, v := range m {
			b := &bytes.Buffer{}
			json.Indent(b, v, "", "  ")
			fmt.Printf("%v\n%s\n\n", k, b)
		}
	}
}
