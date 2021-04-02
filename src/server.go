package main

import (
	"encoding/json"
	"fmt"
	"github.com/SmartsYoung/test/src/api"
	"github.com/widuu/gojson"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

var (
	defaultLabels = []string{"kind", "priority", "area"}
	labelRegex    = regexp.MustCompile(`(?m)^//(comp|sig|bug)\s*(.*?)\s*$`)
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func handlePostJson(writer http.ResponseWriter, request *http.Request) {
	h := request.Header
	for key, value := range h {
		if key == "X-Gitee-Event" {
			str1 := strings.Join(value, "")
			if str1 == "Issue Hook" {
				issue_process(request)
			}
			if str1 == "Note Hook" {
				note_process(request)
			}
		}
	}
}

func getToken() []byte {
	return []byte("adb08695039522366c4a645e1e6a3dd4")
}

func getTokenstr() string {
	return "adb08695039522366c4a645e1e6a3dd4"
}

func note_process(request *http.Request) {
	s, _ := ioutil.ReadAll(request.Body) //把	body 内容读入字符串 s
	str := string(s)
	name := gojson.Json(str).Get("comment").Get("user").Get("name").Tostring()
	noteBody := gojson.Json(str).Get("comment").Get("body").Tostring()
	issue_num := gojson.Json(str).Get("issue").Get("number").Tostring()
	var v interface{}
	json.Unmarshal(s, &v)
	labels := v.(map[string]interface{})["issue"].(map[string]interface{})["labels"].([]interface{})
	label_str := make([]string, 0)
	for _, o := range labels {
		name := o.(map[string]interface{})["name"].(string)
		label_str = append(label_str, name)
	}
	//if name != "test-bot" {
	//	c := api.NewClient(getToken)
	//	res := c.AddIssueAssignee("Meta-OSS", "FenixscanX", issue_num, token, "clement_li")
	//	if res != nil {
	//		fmt.Println(res.Error())
	//		return
	//	}
	//	resc := c.AddIssueLabel("Meta-OSS", "FenixscanX", issue_num, label_str)
	//	if resc != nil {
	//		fmt.Println(resc.Error())
	//		return
	//	}
	//}
	if name != "test-bot" {
		c := api.NewClient(getToken)
		labelMatches := labelRegex.FindAllStringSubmatch(noteBody, -1)
		if len(labelMatches) == 0 {
			return
		}
		//fmt.Println(labelMatches)

		var labelsToAdd []string
		// Get labels to add and labels to remove from regexp matches
		labelsToAdd = getLabelsFromREMatches(labelMatches)

		// Add labels
		//fmt.Println(labelsToAdd)

		resc := c.AddIssueLabel("Meta-OSS", "FenixscanX", issue_num, labelsToAdd)
		if resc != nil {
			fmt.Println(resc.Error())
			return
		}
		return
	}
}

func issue_process(request *http.Request) {
	s, _ := ioutil.ReadAll(request.Body) //把	body 内容读入字符串 s
	str := string(s)

	issue_num := gojson.Json(str).Get("issue").Get("number").Tostring()
	c := api.NewClient(getToken)
	res := c.CreateGiteeIssueComment("Meta-OSS", "FenixscanX", issue_num, "Please add labels, for example, "+
			`if you are filing a runtime issue, you can type "//comp/infra" in comment,`+
			` you can visit "https://shimo.im/sheets/8pKDkqKqdycHRwWV/MODOC/" to find more labels`)
	if res != nil {
		fmt.Println(res.Error())
		return
	}
}

// Get Labels from Regexp matches
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
	http.HandleFunc("/", handlePostJson)
	http.ListenAndServe(":8008", nil)
}
