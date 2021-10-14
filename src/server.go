package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	gitee_utils "gitee.com/lizi/test-bot/src/gitee-utils"
	"gitee.com/openeuler/go-gitee/gitee"
)

var JsonByte []byte
var prComment []byte
var issueComment []byte
var decisionComment []byte
var partiComment []byte
var partiAiComment []byte
var token []byte
var assignComment []byte
var regexpWords = "(mailinglist|maillist|mail|邮件|邮箱|subscribe|订阅" +
	"|etherpad|meetingrecord" +
	"|cla|CLA|signagreement|协议|签署" +
	"|guarding|jenkins|staticcheck|test|compile|robot|测试|编译|检查" +
	"|website|blog|mirror|下载|官网|博客|镜像" +
	"|meeting|会议|例会" +
	"|sensitive|敏感" +
	"|log|日志" +
	"|label|标签" +
	"|access|permission|权限" +
	"|requirement|featurerequest|需求" +
	"|translation|翻译" +
	"|bug|cve|CVE|security" +
	"|gitee|Gitee|Git|git" +
	"|sig|Sig|SIG" +
	"|scheduling|调度" +
	"|OBS|PRM|ISO" +
	"|src-openeuler|openeuler|openEuler)"
var (
	labelRegex = regexp.MustCompile(`(?m)//(mailing|etherpad|CLA|guarding|website|meeting|kind|bug|CVE|security|activity|gitee|git|sig|release|build|repo|)(\S*)`)
	// labelRegexTitle = regexp.MustCompile(`^(.*)(Lite|LITE)\s*(.*?)\s*$`)
	labelRegexWords = regexp.MustCompile(regexpWords)
)

type Mentor struct {
	Dir   string `json:"directory"`
	Label string `json:"label"`
	Name  string `json:"name"`
}

func getToken() []byte {
	return token
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
		go handleCommentEvent(&ic)
	//case "Merge Request Hook":
	//var ip gitee.PullRequestEvent
	//if err := json.Unmarshal(payload, &ip); err != nil {
	//	return
	//}
	//go handlePullRequestEvent(&ip)
	default:
		return
	}
}

func handleIssueEvent(i *gitee.IssueEvent) {
	if *(i.Action) != "open" {
		return
	}
	assignee := ""
	strLabels := ""
	orgOrigin := "all_for_test"
	labelsToAdd_str := ""
	issueNum := i.Issue.Number
	org := i.Repository.Namespace
	repo := i.Repository.Name
	issueBody := i.Issue.Body
	// issueTitle := i.Issue.Title
	issueType := i.Issue.TypeName
	issueInit := i.Issue.Labels
	assigneeInit := i.Issue.Assignee
	issueMaker := i.Issue.User.Login

	// ignore := false
	decision := false
	issueTemp := string(issueComment[:])
	decisionTemp := string(decisionComment[:])
	// partiAiTemp := string(partiAiComment[:])
	assigneeStr := ""

	c := gitee_utils.NewClient(getToken)

	if len(issueInit) == 0 {
		res := c.CreateGiteeIssueComment(org, repo, issueNum, issueTemp)
		if res != nil {
			fmt.Println(res.Error())
			return
		}
		var labelsToAdd []string

		labelMatches := labelRegex.FindAllStringSubmatch(issueBody, -1)
		if len(labelMatches) != 0 {
			labelsToAdd = getLabelsFromREMatches(labelMatches)
		}

		// var labelFindTi []string
		// var nameFindTi []string
		// labelTiMatches := labelRegexTitle.FindAllStringSubmatch(issueTitle, -1)
		// if len(labelTiMatches) != 0 {
		// 	nameFindTi = getLabelsFromBodyMatches(labelTiMatches)
		// }
		// labelFindTi = getLabel(JsonByte, nameFindTi)

		// if len(labelFindTi) != 0 {
		// 	labelsToAdd = append(labelsToAdd, labelFindTi...)
		// }

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
		case "Task-Tracking":
			labelsToAdd = append(labelsToAdd, "kind/task")
			break
		case "任务":
			labelsToAdd = append(labelsToAdd, "kind/task")
			break
		default:
			labelsToAdd = append(labelsToAdd, "kind/bug", "stat/help-wanted")
			break
		}

		if assigneeInit != nil {
			assignee = assigneeInit.Login
		} else if isUserInEnt(issueMaker, orgOrigin, c) {
			assignee = issueMaker
		} else {
			assignee = getLabelAssignee(JsonByte, labelsToAdd)
		}
		assigneeStr = " @" + assignee + " "

		labelsToAdd_str = strings.Join(labelsToAdd, ",")
		rese := c.AssignGiteeIssue(org, repo, labelsToAdd_str, issueNum, assignee)
		if rese != nil {
			fmt.Println(rese.Error())
			return
		}

		issueBody = strings.Replace(issueBody, " ", "", -1)
		issueBody = strings.Replace(issueBody, "\n", "", -1)
		var labelFind []string
		var nameFind []string
		labelBoMatches := labelRegexWords.FindAllStringSubmatch(issueBody, -1)

		if len(labelBoMatches) != 0 {
			for _, tt := range labelBoMatches {
				for _, t := range tt {
					fmt.Printf("labelMatches from noteBody:\n%s \n", t)
				}
			}
			nameFind = getLabelsFromBodyMatches(labelBoMatches)
		}

		if len(nameFind) != 0 {
			fmt.Println("debug message 1:")
			for _, name := range nameFind {
				fmt.Printf("%s ", name)
			}
		}

		labelFind = getLabel(JsonByte, nameFind)

		if len(labelFind) != 0 {
			for _, strLabel := range labelFind {
				strLabels = strLabels + "**//" + strLabel + "**" + "\n"
			}
			partiTemp := string(partiComment[:])
			helloWord := ""
			helloWord = strings.Replace(partiTemp, "{"+"issueMaker"+"}", fmt.Sprintf("%v", issueMaker), -1)

			if assignee != "" {
				helloWord = strings.Replace(helloWord, "{"+"assignee"+"}", fmt.Sprintf("%v", assignee), -1)
			} else {
				helloWord = strings.Replace(helloWord, "@"+"{"+"assignee"+"}", fmt.Sprintf("%v", ""), -1)
			}

			if strings.Contains(strLabels, "good-first-issue") {
				helloWord = strings.Replace(helloWord, "{"+"goodissue"+"}", fmt.Sprintf("%v", ", 因为这个issue看起来是文档类问题, 适合新手开发者解决"), -1)
			} else {
				helloWord = strings.Replace(helloWord, "{"+"goodissue"+"}", fmt.Sprintf("%v", ""), -1)
			}

			helloWord = strings.Replace(helloWord, "{"+"label"+"}", fmt.Sprintf("%v", strLabels), -1)

			resLabel := c.CreateGiteeIssueComment(org, repo, issueNum, helloWord)
			if resLabel != nil {
				fmt.Println(resLabel.Error())
				return
			}
		}
	} else {

		res := c.CreateGiteeIssueComment(org, repo, issueNum, issueTemp)
		if res != nil {
			fmt.Println(res.Error())
			return
		}

		var labelFindTemp []string
		for _, label := range issueInit {
			// if strings.Contains(label.Name, "comp/") ||
			// 	strings.Contains(label.Name, "sig/") ||
			// 	strings.Contains(label.Name, "wg/") {
			// 	ignore = false
			// 	break
			// }

			strings.Join(labelFindTemp, label.Name)
			if label.Name == "kind/decision" {
				decision = true
			}
		}
		if assigneeInit != nil {
			assignee = assigneeInit.Login
		} else if isUserInEnt(issueMaker, orgOrigin, c) {
			assignee = issueMaker
			rese := c.AssignGiteeIssue(org, repo, "", issueNum, assignee)
			if rese != nil {
				fmt.Println(rese.Error())
				return
			}
		} else {
			assignTemp := string(assignComment[:])
			assignee = getLabelAssignee(JsonByte, labelFindTemp)
			assignTemp = strings.Replace(assignTemp, "{"+"assignee"+"}", fmt.Sprintf("%v", assignee), -1)
			rs := c.CreateGiteeIssueComment(org, repo, issueNum, assignTemp)
			if rs != nil {
				fmt.Println(res.Error())
				return
			}
		}

		if decision == true {
			assigneeStr = " @" + assignee + " "
			Temp := "hello, @" + issueMaker + assigneeStr + " " + decisionTemp + "\n"
			res := c.CreateGiteeIssueComment(org, repo, issueNum, Temp)
			if res != nil {
				fmt.Println(res.Error())
				return
			}
		}
	}
}

func handleCommentEvent(i *gitee.NoteEvent) {
	switch *(i.NoteableType) {
	case "Issue":
		go handleIssueCommentEvent(i)
	//case "PullRequest":
	//	go handlePRCommentEvent(i)
	default:
		return
	}
}

func handleIssueCommentEvent(i *gitee.NoteEvent) {
	if *(i.Action) != "comment" {
		return
	}
	assignee := ""
	labelsToAddStr := ""
	org := i.Repository.Namespace
	repo := i.Repository.Name
	name := i.Comment.User.Name
	noteBody := i.Comment.Body
	issueNum := i.Issue.Number
	labels := i.Issue.Labels
	issueMaker := i.Issue.User.Login
	assigneeStr := ""
	decisionTemp := string(decisionComment[:])
	// partiAiTemp := string(partiAiComment[:])
	if i.Issue.Assignee != nil {
		assignee = i.Issue.Assignee.Login
		assigneeStr = " @" + assignee + " "
		fmt.Printf("CODE0:Issue %s was assigned to %s \n", issueNum, assignee)
	}
	labelStrs := make([]string, 0)
	for _, o := range labels {
		labelStrs = append(labelStrs, o.Name)
	}
	if name != "test-bot" {
		c := gitee_utils.NewClient(getToken)
		labelMatches := labelRegex.FindAllStringSubmatch(noteBody, -1)
		if len(labelMatches) == 0 {
			return
		}
		for _, tt := range labelMatches {
			for _, t := range tt {
				fmt.Printf("labelMatches from noteBody:\n%s \n", t)
			}
		}
		var labelsToAdd []string
		labelsToAdd = getLabelsFromREMatches(labelMatches)
		for _, labelTemp := range labelsToAdd {
			fmt.Printf("labelGet from matches: \n%s \n", labelTemp)
		}

		if strings.Contains(noteBody, "good-first-issue") {
			astr := "如果您是第一次贡献社区，可以参考我们的贡献指南：https://gitee.com/openeuler/blob/master/CONTRIBUTING.md"
			res := c.CreateGiteeIssueComment(org, repo, issueNum, astr)
			if res != nil {
				fmt.Println(res.Error())
				return
			}
		}

		if len(labelStrs) != 0 {
			labelsToAdd = append(labelsToAdd, labelStrs...)
		}
		labelsToAddStr = strings.Join(labelsToAdd, ",")

		if assignee == "" {
			assignee = getLabelAssignee(JsonByte, labelsToAdd)
			assigneeStr = " @" + assignee + " "
			fmt.Printf("CODE1:Issue %s was assigned to %s \n", issueNum, assignee)
		}
		rese := c.AssignGiteeIssue(org, repo, labelsToAddStr, issueNum, assignee)
		if rese != nil {
			fmt.Println(rese.Error())
			return
		}

		for _, label := range labelsToAdd {
			if label == "kind/decision" {
				Temp := "hello, @" + issueMaker + assigneeStr + " " + decisionTemp + "\n"
				res := c.CreateGiteeIssueComment(org, repo, issueNum, Temp)
				if res != nil {
					fmt.Println(res.Error())
					return
				}
			}
		}
	}
}

func handlePRCommentEvent(i *gitee.NoteEvent) {
	if *(i.Action) != "comment" {
		return
	}
	prNum := i.PullRequest.Number
	org := i.Repository.Namespace
	repo := i.Repository.Name
	name := i.Comment.User.Name
	noteBody := i.Comment.Body
	if name != "test-bot" {
		c := gitee_utils.NewClient(getToken)
		var labelsToAdd []string
		labelMatches := labelRegex.FindAllStringSubmatch(noteBody, -1)
		if len(labelMatches) != 0 {
			labelsToAdd = getLabelsFromREMatches(labelMatches)
			rese := c.AddPRLabel(org, repo, int(prNum), labelsToAdd)
			if rese != nil {
				fmt.Println(rese.Error())
				return
			}
		}
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
		label := strings.TrimSpace(strings.TrimLeft(match[0], "//"))
		labels = append(labels, label)
	}
	return labels
}

func getLabelsFromBodyMatches(matches [][]string) []string {
	var labels []string
	for _, match := range matches {
		label := strings.ToLower(strings.TrimSpace(match[0]))
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
		for j := range labels {
			if mentors[i].Label == labels[j] {
				return mentors[i].Name
			}
		}
	}
	return ""
}

func getLabel(mentorsJson []byte, dirs []string) []string {
	var mentors []Mentor
	var labels []string
	if err := json.Unmarshal(mentorsJson, &mentors); err != nil {
		fmt.Println(err)
		return labels
	}
	for i := range mentors {
		for j := range dirs {
			if mentors[i].Dir == dirs[j] {
				labels = append(labels, mentors[i].Label)
			}
		}
	}
	return labels
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

func getRecommendation(c gitee_utils.Client, labels []string) string {
	var labelArr []string
	for _, label := range labels {
		labelArr = append(labelArr, label)
	}
	labelStr := strings.Join(labelArr, ",")
	participants, res := c.GetRecommendation(labelStr)
	if res != nil {
		fmt.Println(res.Error())
		return ""
	}
	return participants
}

func loadFile(path, fileType string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		defer jsonFile.Close()
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	switch {
	case fileType == "json":
		JsonByte = byteValue
	case fileType == "issue":
		issueComment = byteValue
	case fileType == "pr":
		prComment = byteValue
	case fileType == "decision":
		decisionComment = byteValue
	case fileType == "parti":
		partiComment = byteValue
	case fileType == "partiAI":
		partiAiComment = byteValue
	case fileType == "token":
		token = byteValue
	case fileType == "assign":
		assignComment = byteValue
	default:
		fmt.Printf("no filetype\n")
	}
	return nil
}

func configFile() {
	loadFile("src/data/mentor.json", "json")
	loadFile("src/data/issueComTemplate.md", "issue")
	loadFile("src/data/decisionTemplate.md", "decision")
	loadFile("src/data/prComTemplate.md", "pr")
	loadFile("src/data/partiTemplate.md", "parti")
	loadFile("src/data/partiTemplate_ai.md", "partiAI")
	loadFile("src/data/token.md", "token")
	loadFile("src/data/assignTemplate.md", "assign")
}

func main() {
	configFile()
	http.HandleFunc("/", ServeHTTP)
	http.ListenAndServe(":8008", nil)
}
