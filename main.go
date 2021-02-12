package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/mod/modfile"
)

func main() {
	b, err := ioutil.ReadFile("modreplace.txt")
	if err != nil {
		log.Panicf("failed to read modreplace.txt: %v", err)
	}
	paths := strings.Split(strings.TrimSpace(string(b)), "\n")

	b, err = ioutil.ReadFile("go.mod")
	if err != nil {
		log.Panicf("failed to read go.mod: %v", err)
	}

	modFile, err := modfile.Parse("go.mod", b, nil)
	if err != nil {
		log.Panicf("failed to parse go.mod: %v", err)
	}

	Replace(paths, modFile)
	b, err = modFile.Format()
	if err != nil {
		log.Panicf("failed to format go.mod: %v", err)
	}

	ioutil.WriteFile("go.mod", b, os.ModePerm)
}
