package main

import (
	bytes2 "bytes"
	"encoding/gob"
	"fmt"
	"github.com/looplanguage/compiler/compiler"
	"github.com/looplanguage/loop/lexer"
	"github.com/looplanguage/loop/parser"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("not enough arguments. specify filename to compile.")
		return
	}

	bytes, err := ioutil.ReadFile(os.Args[1])

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(fmt.Sprintf("compiling %q", os.Args[1]))

	fileContent := string(bytes)

	l := lexer.Create(fileContent)
	p := parser.Create(l)
	program := p.Parse()

	if len(p.Errors) != 0 {
		for _, err := range p.Errors {
			fmt.Println(err)
		}
		return
	}

	comp := compiler.Create()
	err = comp.Compile(program)

	if err != nil {
		log.Fatalln(err)
	}

	compiler.RegisterGobTypes()

	var constantBytes bytes2.Buffer

	enc := gob.NewEncoder(&constantBytes)
	err = enc.Encode(comp.Bytecode())

	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(comp.Bytecode().Instructions.String())

	ioutil.WriteFile(filepath.Dir(os.Args[1])+"/"+fileNameWithoutExtension(filepath.Base(os.Args[1]))+".lpx", constantBytes.Bytes(), 0644)
}

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
