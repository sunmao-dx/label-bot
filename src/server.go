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
var prComment []byte
var issueComment []byte
var decisionComment []byte
var partiComment []byte
var partiAiComment []byte

var (
	labelRegex    = regexp.MustCompile(`(?m)^//(comp|sig|good|bug|wg|stat|kind|device|env|ci|mindspore|DFX|usability|users|0|1|2)\s*(.*?)\s*$`)
	labelRegexTitle    = regexp.MustCompile(`^(.*)(Lite|LITE)\s*(.*?)\s*$`)
	labelRegexBody    = regexp.MustCompile(`^(.*)(/ops/|/kernel/|/minddata/|/parallel/|/optimizer/|/pynative/|/kernel_compiler/|/device/|/parse/|/cxx_api/|/debug/|/ps/|/pybind_api/|/transform/|/vm/|/communication/|/dataset/|/lite/|/mindrecord/|/nn/|/profiler/|/train/|/model_zoo/|/akg/)\s*(.*?)\s*$`)
)

type Mentor struct {
	Dir string `json:"directory"`
	Label string `json:"label"`
	Name  string `json:"name"`
}

func getToken() []byte {
	return []byte("adb08695039522366c4a645e1e6a3dd4")
}

func getMLToken() string {
	return "eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNzkwNDc3OSwiZXhwIjoxNjI3OTA4Mzc5fQ.eyJpZCI6Imxva2kifQ.lwmt3yKzmqVdzHdEcFWQxuSw_g40JOBPiMZHIOEmybJTiumkL0G3U7x04MqlKWJFFUWsA4RjZt0UuAffocBddw"
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
	case "Merge Request Hook":
		var ip gitee.PullRequestEvent
		if err := json.Unmarshal(payload, &ip); err != nil {
			return
		}
		go handlePullRequestEvent(&ip)
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
	orgOrigin := "mind_spore"
	labelsToAdd_str := ""
	issueNum := i.Issue.Number
	org := i.Repository.Namespace
	repo := i.Repository.Name
	issueBody := i.Issue.Body
	issueTitle := i.Issue.Title
	issueType := i.Issue.TypeName
	issueInit := i.Issue.Labels
    assigneeInit := i.Issue.Assignee
	issueMaker := i.Issue.User.Login

	ignore := false
	decision := false
	issueTemp := string(issueComment[:])
	decisionTemp := string(decisionComment[:])
	partiAiTemp := string(partiAiComment[:])
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

		issueBody = strings.Replace(issueBody, " ", "", -1)
		issueBody = strings.Replace(issueBody, "\n", "", -1)
		var labelFind []string
		var nameFind []string
		labelBoMatches := labelRegexBody.FindAllStringSubmatch(issueBody, -1)
		if len(labelBoMatches) != 0 {
			nameFind = getLabelsFromBodyMatches(labelBoMatches)
		}
		labelFind = getLabel(JsonByte, nameFind)

		var labelFindTi []string
		var nameFindTi []string
		labelTiMatches := labelRegexTitle.FindAllStringSubmatch(issueTitle, -1)
		if len(labelTiMatches) != 0 {
			nameFindTi = getLabelsFromBodyMatches(labelTiMatches)
		}
		labelFindTi = getLabel(JsonByte, nameFindTi)

		if len(labelFindTi) != 0 {
			labelsToAdd = append(labelsToAdd, labelFindTi...)
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
		assignee = getLabelAssignee(JsonByte, labelsToAdd)
		if isUserInEnt(issueMaker, orgOrigin, c) {
			assignee = issueMaker
		}
		if assigneeInit != nil {
			assignee = assigneeInit.Login
			assigneeStr = " @" + assignee + " "
		}
		labelsToAdd_str = strings.Join(labelsToAdd,",")
		rese := c.AssignGiteeIssue(org, repo, labelsToAdd_str, issueNum, assignee)
		if rese != nil {
			fmt.Println(rese.Error())
			return
		}

		if len(labelFind) != 0 {
			for _, strLabel := range labelFind {
				strLabels = strLabels + "//" + strLabel + "\n"
			}
			partiTemp := string(partiComment[:])
			helloWord := ""
			helloWord = strings.Replace(partiTemp, "{"+"issueMaker"+"}", fmt.Sprintf("%v", issueMaker), -1)
			helloWord = strings.Replace(helloWord, "{"+"assignee"+"}", fmt.Sprintf("%v", assignee), -1)

			labelWord := helloWord + strLabels
			resLabel := c.CreateGiteeIssueComment(org, repo, issueNum, labelWord)
			if resLabel != nil {
				fmt.Println(resLabel.Error())
				return
			}
		}
	} else {
		if assigneeInit != nil {
			assignee = assigneeInit.Login
			assigneeStr = " @" + assignee + " "
		}
		for _, label:= range issueInit {
			if strings.Contains(label.Name,"comp/") ||
				strings.Contains(label.Name,"sig/") ||
				strings.Contains(label.Name,"wg/") {
				ignore = true
				break
			}
			if label.Name == "kind/decision" {
				decision = true
				break
			}
		}
		if ignore == false {
			res := c.CreateGiteeIssueComment(org, repo, issueNum, issueTemp)
			if res != nil {
				fmt.Println(res.Error())
				return
			}
		} else {
			if repo != "community"{
				var labelArr []string
				for _, label:= range issueInit {
					labelArr = append(labelArr, label.Name)
				}
				participants := getRecommendation(c, labelArr)
				if participants == "" {
					return
				}
				partiArr := strings.Split(participants, ",")
				issueAssignee := partiArr[0]
				var coAssigneeToAdd []string
				coAssigneeToAdd = append(coAssigneeToAdd, partiArr[1:]...)
				coAssignee := strings.Join(coAssigneeToAdd, ",")
				participantsStr := strings.Replace(partiAiTemp, "{"+"issueMaker"+"}", fmt.Sprintf("%v", issueMaker), -1)
				participantsStr = strings.Replace(participantsStr, "{"+"issueAssignee"+"}", fmt.Sprintf("%v", issueAssignee), -1)
				participantsStr = strings.Replace(participantsStr, "{"+"issueCoAssignee"+"}", fmt.Sprintf("%v", coAssignee), -1)
				res := c.CreateGiteeIssueComment(org, repo, issueNum, participantsStr)
				if res != nil {
					fmt.Println(res.Error())
					return
				}
			}
		}
		if decision == true {
			Temp := "hello, @" + issueMaker + assigneeStr + " " + decisionTemp + "\n"
			res := c.CreateGiteeIssueComment(org, repo, issueNum, Temp)
			if res != nil {
				fmt.Println(res.Error())
				return
			}
		}
	}
}

func handlePullRequestEvent(i *gitee.PullRequestEvent) {
	if *(i.Action) != "open" {
		return
	}
	prNum := i.PullRequest.Number
	org := i.Repository.Namespace
	repo := i.Repository.Name
	prBody := i.PullRequest.Body
	prTemp := string(prComment[:])
	c := gitee_utils.NewClient(getToken)
	res := c.CreatePRComment(org, repo, int(prNum), prTemp)
	if res != nil {
		fmt.Println(res.Error())
		return
	}

	var labelsToAdd []string
	labelMatches := labelRegex.FindAllStringSubmatch(prBody, -1)
	if len(labelMatches) != 0 {
		labelsToAdd = getLabelsFromREMatches(labelMatches)
		rese := c.AddPRLabel(org, repo, int(prNum), labelsToAdd)
		if rese != nil {
			fmt.Println(rese.Error())
			return
		}
	}
}

func handleCommentEvent(i *gitee.NoteEvent) {
	switch *(i.NoteableType) {
	case "Issue":
		go handleIssueCommentEvent(i)
	case "PullRequest":
		go handlePRCommentEvent(i)
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
	partiAiTemp := string(partiAiComment[:])
	if i.Issue.Assignee != nil{
		assignee = i.Issue.Assignee.Login
		assigneeStr = " @" + assignee + " "
		fmt.Println(assignee)
	}
	labelStrs := make([]string, 0)
	for _, o := range labels {
		nameStr := o.Name
		labelStrs = append(labelStrs, nameStr)
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
			if len(labelStrs) != 0{
				labelsToAdd = append(labelsToAdd, labelStrs...)
			}
			labelsToAdd = append(labelsToAdd, labelStrs...)
			labelsToAddStr = strings.Join(labelsToAdd,",")
			resd := c.AssignGiteeIssue(org, repo, labelsToAddStr, issueNum, assignee)
			if resd != nil {
				fmt.Println(resd.Error())
				return
			}
			for _, label:= range labelsToAdd {
				if label == "kind/decision" {
					Temp := "hello, @" + issueMaker + assigneeStr + " " + decisionTemp + "\n"
					res := c.CreateGiteeIssueComment(org, repo, issueNum, Temp)
					if res != nil {
						fmt.Println(res.Error())
						return
					}
				}
			}
			return
		}
		assignee = getLabelAssignee(JsonByte, labelsToAdd)
		if len(labelStrs) != 0{
			labelsToAdd = append(labelsToAdd, labelStrs...)
		}
		labelsToAddStr = strings.Join(labelsToAdd,",")

		if repo != "community"{
			participants := getRecommendation(c, labelsToAdd)
			if participants == "" {
				return
			}
			partiArr := strings.Split(participants, ",")
			issueAssignee := partiArr[0]
			var coAssigneeToAdd []string
			coAssigneeToAdd = append(coAssigneeToAdd, partiArr[1:]...)
			coAssignee := strings.Join(coAssigneeToAdd, ",")
			participantsStr := strings.Replace(partiAiTemp, "{"+"issueMaker"+"}", fmt.Sprintf("%v", issueMaker), -1)
			participantsStr = strings.Replace(participantsStr, "{"+"issueAssignee"+"}", fmt.Sprintf("%v", issueAssignee), -1)
			participantsStr = strings.Replace(participantsStr, "{"+"issueCoAssignee"+"}", fmt.Sprintf("%v", coAssignee), -1)
			res := c.CreateGiteeIssueComment(org, repo, issueNum, participantsStr)
			if res != nil {
				fmt.Println(res.Error())
				return
			}
		}

		rese := c.AssignGiteeIssue(org, repo, labelsToAddStr, issueNum, assignee)
		if rese != nil {
			fmt.Println(rese.Error())
			return
		}
		for _, label:= range labelsToAdd {
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
	if name != "mindspore-dx-bot" {
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
		label := strings.TrimSpace(strings.TrimLeft(match[0],"//"))
		labels = append(labels, label)
	}
	return labels
}

func getLabelsFromBodyMatches(matches [][]string) []string {
	var labels []string
	for _, match := range matches {
		label := strings.ToLower(strings.TrimSpace(strings.Trim(match[2],"/")))
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

func getLabel(mentorsJson []byte, dirs []string) []string {
	var mentors []Mentor
	var labels []string
	if err := json.Unmarshal(mentorsJson, &mentors); err != nil {
		fmt.Println(err)
		return labels
	}
	for i := range mentors {
		for j := range dirs{
			if mentors[i].Dir == dirs[j]{
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
	for _, label:= range labels {
		labelArr = append(labelArr, label)
	}
	labelStr := strings.Join(labelArr,",")
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
	case fileType == "json" :
		JsonByte = byteValue
	case fileType == "issue" :
		issueComment = byteValue
	case fileType == "pr":
		prComment = byteValue
	case fileType == "decision" :
		decisionComment = byteValue
	case fileType == "parti" :
		partiComment = byteValue
	case fileType == "partiAI" :
		partiAiComment = byteValue
	default:
		fmt.Printf("no filetype\n" )
	}
	return nil
}

func main() {
	loadFile("src/data/mentor.json", "json")
	loadFile("src/data/issueComTemplate.md", "issue")
	loadFile("src/data/decisionTemplate.md", "decision")
	loadFile("src/data/prComTemplate.md", "pr")
	loadFile("src/data/partiTemplate.md", "parti")
	loadFile("src/data/partiTemplate_ai.md", "partiAI")
	http.HandleFunc("/", ServeHTTP)
	http.ListenAndServe(":8008", nil)
}
