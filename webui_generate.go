// +build ignore

package main

import (
	"log"

	"wab"

	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(wab.WebUI, vfsgen.Options{
		PackageName:  "wab",
		BuildTags:    "!dev",
		VariableName: "WebUI",
		Filename:     "webui.go",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
