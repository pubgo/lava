package main

import (
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"os"
)

func main() {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)

	md := []byte(

		`

"## markdown document"

Name    | Age
--------|------
Bob     ||
Alice   | 23
========|======
Total   | 23

Cat
: Fluffy animal everyone likes

Internet
: Vector of transmission for pictures of cats


This is a footnote.[^1]

[^1]: the footnote text.


$$
\left[ \begin{array}{a} a^l_1 \\ ⋮ \\ a^l_{d_l} \end{array}\right]
= \sigma(
 \left[ \begin{matrix}
 	w^l_{1,1} & ⋯  & w^l_{1,d_{l-1}} \\
 	⋮ & ⋱  & ⋮  \\
 	w^l_{d_l,1} & ⋯  & w^l_{d_l,d_{l-1}} \\
 \end{matrix}\right]  ·
 \left[ \begin{array}{x} a^{l-1}_1 \\ ⋮ \\ ⋮ \\ a^{l-1}_{d_{l-1}} \end{array}\right] +
 \left[ \begin{array}{b} b^l_1 \\ ⋮ \\ b^l_{d_l} \end{array}\right])
 $$

{#id3 .myclass fontsize="tiny"}
# Header 1
`,
	)
	//html := markdown.ToHTML(md, parser, nil)
	//fmt.Println(string(html))

	var node = parser.Parse(md)
	ast.Print(os.Stdout, node)
}
