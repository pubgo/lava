package prompt


import prompt "github.com/c-bata/go-prompt"

func init() {
	p := prompt.New(
		nil,
		nil,
		prompt.OptionTitle("kube-prompt: interactive kubernetes client"),
		prompt.OptionPrefix(">>> "),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionCompletionWordSeparator(""),
	)
	p.Run()
}
