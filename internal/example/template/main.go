package main

import (
	"github.com/open2b/scriggo"
	"github.com/open2b/scriggo/builtin"
	"github.com/open2b/scriggo/native"
	"log"
	"os"
)

func main() {
	// Content of the template file to run.
	content := []byte(`
    <!DOCTYPE html>
    <html>
    <head>Hello</head> 
    <body>
        {% who := "World" %}
        Hello, {{ replace(who, "World", "世界", 1) }}!
    </body>
    </html>
    `)

	// Create a file system with the file of the template to run.
	fsys := scriggo.Files{"index.html": content}

	// Allow to use the "replace" built-in in the template file.
	globals := native.Declarations{
		"replace": builtin.Replace,
	}
	opts := scriggo.BuildOptions{Globals: globals}

	// Build the template.
	template, err := scriggo.BuildTemplate(fsys, "index.html", &opts)
	if err != nil {
		log.Fatal(err)
	}

	// Run the template and print it to the standard output.
	err = template.Run(os.Stdout, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

}
