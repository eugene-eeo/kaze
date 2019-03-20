package main

import "bytes"
import "strings"
import "os/exec"

func execUserSelector(command []string, chunks []string) string {
	out := bytes.Buffer{}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = strings.NewReader(strings.Join(chunks, "\n"))
	cmd.Stdout = &out
	cmd.Run()
	// Command success => some entry selected
	if !cmd.ProcessState.Success() {
		return ""
	}
	s := string(out.Bytes())
	s = strings.TrimRight(s, "\n")
	return s
}

func execMixedSelector(actions []NotificationAction, links []Hyperlink, actionsCallback func(string)) {
	chunks := make([]string, len(actions)+len(links))
	i := 0
	for _, action := range actions {
		chunks[i] = "action\t" + action.Key + "\t" + action.Value
		i++
	}
	for _, link := range links {
		chunks[i] = "link\t" + link.Text + "\t" + link.Href
		i++
	}
	choice := execUserSelector([]string{"slouch", "pipe"}, chunks)
	selection := strings.SplitN(choice, "\t", 3)
	if len(selection) == 3 {
		switch selection[0] {
		case "action":
			actionsCallback(selection[1])
		case "link":
			exec.Command("xdg-open", selection[2]).Run()
		}
	}
}