package main

import (
	"github.com/webview/webview"

	"log"
)

func Example() {
	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("Hello")
	w.SetSize(800, 600, webview.HintNone)
	w.Bind("noop", func() string {
		log.Println("hello")
		return "hello"
	})
	w.Bind("add", func(a, b int) int {
		return a + b
	})
	w.Bind("quit", func() {
		w.Terminate()
	})
	w.Navigate(`data:text/html,
		<!doctype html>
		<html>
			<body>hello</body>
			<script>
				window.onload = function() {
					document.body.innerText = ` + "`hello, ${navigator.userAgent}`" + `;
					noop().then(function(res) {
						console.log('noop res', res);
						add(1, 2).then(function(res) {
							console.log('add res', res);
							//quit();
						});
					});
				};
			</script>
		</html>
	)`)
	w.Run()
}

func main() {
	Example()
}
