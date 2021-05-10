package main

import (
	"encoding/json"
	"fmt"
	gitee_utils "gitee.com/lizi/test-bot/src/gitee-utils"
	"gitee.com/openeuler/go-gitee/gitee"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var JsonByte []byte

var (
	labelRegex    = regexp.MustCompile(`(?m)^//(comp|sig|good|bug|wg|stat|kind|device|env|ci|mindspore|DFX|usability|0|1|2)\s*(.*?)\s*$`)
	labelRegexInit    = regexp.MustCompile(`(?m)^//(comp|sig)\s*(.*?)\s*$`)
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
	case "Merge Request Hook":
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
	assignee := ""
	orgOrigin := "mind_spore"
	labelsToAdd_str := ""
	issueNum := i.Issue.Number
	org := i.Repository.Namespace
	repo := i.Repository.Name
	issueBody := i.Issue.Body
	issueType := i.Issue.TypeName
	labelsInit := i.Issue.Labels
	issueMaker := i.Issue.User.Login

	c := gitee_utils.NewClient(getToken)
	if len(labelsInit) == 0{
		res := c.CreateGiteeIssueComment(org, repo, issueNum, "Please add labels (comp or sig), "+
			` also you can visit "https://gitee.com/mindspore/community/blob/master/sigs/dx/docs/labels.md" to find more.`+ "\n" +
			` 为了让问题更快得到响应，请您为该issue打上组件(comp)或兴趣组(sig)标签，打上标签的问题可以直接推送给责任人进行处理。更多的标签可以查看`+
			`https://gitee.com/mindspore/community/blob/master/sigs/dx/docs/labels.md"`+ "\n" +
			` 以组件问题为例，如果你发现问题是data组件造成的，你可以这样评论：`+ "\n" +
			` //comp/data`+ "\n" +
			` 当然你也可以向data SIG组求助，可以这样写：`+ "\n" +
			` //comp/data`+ "\n" +
			` //sig/data`+ "\n" +
			` 如果是一个简单的问题，你可以留给刚进入社区的小伙伴来回答，这时候你可以这样写：`+ "\n" +
			` //good-first-issue`+ "\n" +
			` 恭喜你，你已经学会了使用命令来打标签，接下来就在下面的评论里打上标签吧！`)
		if res != nil {
			fmt.Println(res.Error())
			return
		}
	} else {
		return
	}

	var labelsToAdd []string
	labelMatches := labelRegex.FindAllStringSubmatch(issueBody, -1)
	if len(labelMatches) != 0 {
		labelsToAdd = getLabelsFromREMatches(labelMatches)
	}
	switch issueType {
	case "Bug-Report":
		labelsToAdd = append(labelsToAdd, "kind/bug")
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
	assignee = getLabelAssignee(JsonByte, labelsToAdd)
	if isUserInEnt(issueMaker, orgOrigin, c) {
		assignee = issueMaker
	}
	labelsToAdd_str = strings.Join(labelsToAdd,",")
	rese := c.AssignGiteeIssue(org, repo, labelsToAdd_str, issueNum, assignee)
	if rese != nil {
		fmt.Println(rese.Error())
		return
	}
	return
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
		assignee = i.Issue.Assignee.Login
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
			if len(label_strs) != 0{
				labelsToAdd = append(labelsToAdd, label_strs...)
			}
			labelsToAdd = append(labelsToAdd, label_strs...)
			labelsToAdd_str = strings.Join(labelsToAdd,",")
			resd := c.AssignGiteeIssue(org, repo, labelsToAdd_str, issue_num, assignee)
			if resd != nil {
				fmt.Println(resd.Error())
				return
			}
			return
		}
		assignee = getLabelAssignee(JsonByte, labelsToAdd)
		if len(label_strs) != 0{
			labelsToAdd = append(labelsToAdd, label_strs...)
		}
		labelsToAdd_str = strings.Join(labelsToAdd,",")
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

func getLabelsFromREMatches(matches [][]string) []string {
	var labels []string
	for _, match := range matches {
		label := strings.TrimSpace(strings.Trim(match[0],"//"))
		labels = append(labels, label)
	}
	return labels
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

func isUserInOrg(login, orgOrigin string, c gitee_utils.Client) bool {
	orgs, err := c.GetUserOrg(login)
	if err != nil {
		fmt.Println(err)
		return false
	}
	for _, org := range orgs {
		if org.Login == orgOrigin {
			return true
		}
	}
	return false
}

func isUserInEnt(login, entOrigin string, c gitee_utils.Client) bool {
	_, err := c.GetUserEnt(entOrigin, login)
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		return true
	}
}

func loadJson() error {
	jsonFile, err := os.Open("src/data/mentor.json")
	if err != nil {
		fmt.Println(err)
		defer jsonFile.Close()
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	JsonByte = byteValue
	return nil
}

func main() {
	loadJson()
	http.HandleFunc("/", ServeHTTP)
	http.ListenAndServe(":8008", nil)
}
