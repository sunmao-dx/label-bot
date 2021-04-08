package main

import (
	"encoding/json"
	"fmt"
	"gitee.com/openeuler/go-gitee/gitee"
	"github.com/SmartsYoung/test/src/api"
	gitee_utils "github.com/SmartsYoung/test/src/gitee-utils"
	"net/http"
	"regexp"
	"strings"
)

var (
	defaultLabels = []string{"kind", "priority", "area"}
	labelRegex    = regexp.MustCompile(`(?m)^//(comp|sig|bug)\s*(.*?)\s*$`)
)

func getToken() []byte {
	return []byte("adb08695039522366c4a645e1e6a3dd4")
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

// ServeHTTP validates an incoming webhook and puts it into the event channel.
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Event.")
	eventType, _, payload, ok, _ := gitee_utils.ValidateWebhook(w, r)
	if !ok {
		return
	}
	fmt.Fprint(w, "Event received. Have a nice day.")

	switch eventType {
	case "issues":
		var ie gitee.IssueEvent
		if err := json.Unmarshal(payload, &ie); err != nil {
			return
		}
		if err := checkRepository(payload, ie.Repository); err != nil {
			return
		}
		go handleIssueEvent(&ie)
	case "issue_comment":
		var ic gitee.NoteEvent
		if err := json.Unmarshal(payload, &ic); err != nil {
			return
		}
		go handleIssueCommentEvent(&ic)
	default:
		return
	}
}

func handleIssueEvent(i *gitee.IssueEvent) {
	if *(i.Action) != "open" {
		return
	}
	issue_num := i.Issue.Number
	org := i.Repository.Namespace
	repo := i.Repository.Name
	c := api.NewClient(getToken)
	res := c.CreateGiteeIssueComment(org, repo, issue_num, "Please add labels, for example, "+
			`if you are filing a runtime issue, you can type "//comp/runtime" in comment,`+
			` also you can visit "https://gitee.com/mindspore/community/blob/master/sigs/dx/docs/labels.md" to find more labels`)
	if res != nil {
		fmt.Println(res.Error())
		return
	}
}

func handleIssueCommentEvent(i *gitee.NoteEvent) {
	if *(i.NoteableType) != "Issue"{
		return
	}
	if *(i.Action) != "comment" {
		return
	}
	org := i.Repository.Namespace
	repo := i.Repository.Name
	name := i.Comment.User.Name
	noteBody := i.Comment.Body
	issue_num := i.Issue.Number
	labels := i.Issue.Labels
	label_str := make([]string, 0)
	for _, o := range labels {
		name := o.Name
		label_str = append(label_str, name)
	}
	if name != "test-bot" {
		c := gitee_utils.NewClient(getToken)
		labelMatches := labelRegex.FindAllStringSubmatch(noteBody, -1)
		if len(labelMatches) == 0 {
			return
		}
		var labelsToAdd []string
		labelsToAdd = getLabelsFromREMatches(labelMatches)
		resc := c.AddIssueLabel(org, repo, issue_num, labelsToAdd)
		if resc != nil {
			fmt.Println(resc.Error())
			return
		}
		return
	}
}

func checkRepository(payload []byte, rep *gitee.ProjectHook) error {
	if rep == nil {
		return fmt.Errorf("event repository is empty,payload: %s", string(payload))
	}
	return nil
}

func getLabelsFromREMatches(matches [][]string) (labels []string) {
	for _, match := range matches {
		label := strings.TrimSpace(strings.Trim(match[0],"//"))
		labels = append(labels, label)
		fmt.Println(label)
	}
	fmt.Printf("%#v", labels)
	return
}

func main() {
	http.HandleFunc("/", HelloHandler)
	http.ListenAndServe(":8009", nil)
}
