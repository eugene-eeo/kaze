package main

import "bytes"
import "strings"
import "os/exec"

func execUserSelector(chunks []string) string {
	out := bytes.Buffer{}
	cmd := exec.Command("slouch", "pipe")
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

func execActionsSelect(actions []NotificationAction) string {
	chunks := make([]string, len(actions))
	for i, action := range actions {
		chunks[i] = action.Key + "\t" + action.Value
	}
	return strings.SplitN(execUserSelector(chunks), "\t", 2)[0]
}

func execLinkSelect(links []Hyperlink) {
	chunks := make([]string, len(links))
	for i, link := range links {
		chunks[i] = link.Text + "\t" + link.Href
	}
	link := strings.SplitN(execUserSelector(chunks), "\t", 2)[1]
	exec.Command("xdg-open", link).Run()
}
