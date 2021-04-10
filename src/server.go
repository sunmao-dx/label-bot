package main

import (
	"encoding/json"
	"fmt"
	"gitee.com/openeuler/go-gitee/gitee"
	gitee_utils "github.com/SmartsYoung/test/src/gitee-utils"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var jsonByte []byte

var (
	defaultLabels = []string{"kind", "priority", "area"}
	labelRegex    = regexp.MustCompile(`(?m)^//(comp|sig|bug|stat|kind|device|env|ci|0|1|2)\s*(.*?)\s*$`)
)

type Mentor struct {
	Label string `json:"label"`
	Name  string `json:"name"`
}

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

	switch eventType {
	case "Issue Hook":
		var ie gitee.IssueEvent
		if err := json.Unmarshal(payload, &ie); err != nil {
			return
		}
		if err := checkRepository(payload, ie.Repository); err != nil {
			return
		}
		go handleIssueEvent(&ie)
	case "Note Hook":
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
	issueNum := i.Issue.Number
	org := i.Repository.Namespace
	repo := i.Repository.Name
	issueBody := i.Issue.Body
	issueType := i.Issue.TypeName
	c := gitee_utils.NewClient(getToken)
	res := c.CreateGiteeIssueComment(org, repo, issueNum, "Please add labels, for example, "+
			`if you found an issue in data component, you can type "//comp/data" in comment,`+
			` also you can visit "https://gitee.com/mindspore/community/blob/master/sigs/dx/docs/labels.md" to find more.`+ "\n" +
		` 请为该issue打上标签，例如，当你遇到有关data组件的问题时，你可以在评论中输入 "//comp/data"， 这样issue会被打上"comp/data"标签，更多的标签可以查看`+
		`https://gitee.com/mindspore/community/blob/master/sigs/dx/docs/labels.md"`)
	if res != nil {
		fmt.Println(res.Error())
		return
	}
	var labelsToAdd []string
	labelMatches := labelRegex.FindAllStringSubmatch(issueBody, -1)
	if len(labelMatches) != 0 {
		labelsToAdd = getLabelsFromREMatches(labelMatches)
	}
	switch issueType {
	case "Bug-Report":
		labelsToAdd = append(labelsToAdd, "kind/bug", "stat/help-wanted")
		break
	case "RFC":
		labelsToAdd = append(labelsToAdd, "kind/feature", "stat/wait-response")
		break
	case "Requirement":
		labelsToAdd = append(labelsToAdd, "kind/feature", "stat/wait-response")
		break
	case "Empty-Template":
		labelsToAdd = append(labelsToAdd, "stat/wait-response")
		break
	case "Task":
		labelsToAdd = append(labelsToAdd, "kind/task")
		break
	case "任务":
		labelsToAdd = append(labelsToAdd, "kind/task")
		break
	default:
		labelsToAdd = append(labelsToAdd, "kind/bug", "stat/help-wanted")
		break
	}
	resc := c.AddIssueLabel(org, repo, issueNum, labelsToAdd)
	if resc != nil {
		fmt.Println(resc.Error())
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
	assignee := ""
	labelsToAdd_str := ""
	org := i.Repository.Namespace
	repo := i.Repository.Name
	name := i.Comment.User.Name
	noteBody := i.Comment.Body
	issue_num := i.Issue.Number
	labels := i.Issue.Labels
	if i.Issue.Assignee != nil{
		assignee = i.Issue.Assignee.Name
	}
	label_strs := make([]string, 0)
	for _, o := range labels {
		name := o.Name
		label_strs = append(label_strs, name)
	}
	if name != "mindspore-dx-bot" {
		c := gitee_utils.NewClient(getToken)
		labelMatches := labelRegex.FindAllStringSubmatch(noteBody, -1)
		if len(labelMatches) == 0 {
			return
		}
		var labelsToAdd []string
		labelsToAdd = getLabelsFromREMatches(labelMatches)
		if assignee != "" {
			labelsToAdd = append(labelsToAdd, label_strs...)
			labelsToAdd_str = strings.Join(labelsToAdd,",")
			resd := c.AssignGiteeIssue(org, repo, labelsToAdd_str, issue_num, assignee)
			if resd != nil {
				fmt.Println(resd.Error())
				return
			}
			return
		}
		assignee = getLabelAssignee(jsonByte, labelsToAdd)
		if len(label_strs) != 0{
			labelsToAdd = append(labelsToAdd, label_strs...)
			labelsToAdd_str = strings.Join(labelsToAdd,",")
		}
		rese := c.AssignGiteeIssue(org, repo, labelsToAdd_str, issue_num, assignee)
		if rese != nil {
			fmt.Println(rese.Error())
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
	}
	return
}

func getLabelAssignee(mentorsJson []byte, labels []string) string {
	var mentors []Mentor
	if err := json.Unmarshal(mentorsJson, &mentors); err != nil {
		fmt.Println(err)
		return ""
	}
	for i := range mentors {
		for j := range labels{
			if mentors[i].Label == labels[j]{
				return mentors[i].Name
			}
		}
	}
	return ""
}

func loadJson() error {
	jsonFile, err := os.Open("data/mentor.json")
	if err != nil {
		fmt.Println(err)
		defer jsonFile.Close()
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	jsonByte = byteValue
	return nil
}

func main() {
	loadJson()
	http.HandleFunc("/", ServeHTTP)
	http.ListenAndServe(":8008", nil)
}
