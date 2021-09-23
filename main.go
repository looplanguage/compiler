package main

import (
	bytes2 "bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/looplanguage/compiler/compiler"
	"github.com/looplanguage/loop/lexer"
	"github.com/looplanguage/loop/parser"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func main() {
	debugPtr := flag.Bool("debug", false, "Enables printing of bytecode")
	flag.Parse()

	file := flag.Arg(0)

	bytes, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(fmt.Sprintf("compiling %q", file))

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

	dir := filepath.Dir(file) + "/" + fileNameWithoutExtension(filepath.Base(file)) + ".lpx"

	if *debugPtr {
		fmt.Println(comp.Bytecode().Instructions.String())
	}

	err = ioutil.WriteFile(dir, constantBytes.Bytes(), 0644)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("successfully compiled %q to %q", file, dir))
}

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
