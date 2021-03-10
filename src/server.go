package main

import (
	"encoding/json"
	"fmt"
	"github.com/SmartsYoung/test/src/api"
	"github.com/widuu/gojson"
	"io/ioutil"
	"net/http"
	"strings"
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
	token := getTokenstr()
	s, _ := ioutil.ReadAll(request.Body) //把	body 内容读入字符串 s
	str := string(s)
	name := gojson.Json(str).Get("comment").Get("user").Get("name").Tostring()
	issue_num := gojson.Json(str).Get("issue").Get("number").Tostring()
	var v interface{}
	json.Unmarshal(s, &v)
	labels := v.(map[string]interface{})["issue"].(map[string]interface{})["labels"].([]interface{})
	label_str := make([]string, 0)
	for _, o := range labels {
		name := o.(map[string]interface{})["name"].(string)
		label_str = append(label_str, name)
	}
	if name != "test-bot" {
		c := api.NewClient(getToken)
		res := c.AddIssueAssignee("Meta-OSS", "FenixscanX", issue_num, token, "clement_li")
		if res != nil {
			fmt.Println(res.Error())
			return
		}
		resc := c.AddIssueLabel("Meta-OSS", "FenixscanX", issue_num, label_str)
		if resc != nil {
			fmt.Println(resc.Error())
			return
		}
	}
}

func issue_process(request *http.Request) {
	s, _ := ioutil.ReadAll(request.Body) //把	body 内容读入字符串 s
	str := string(s)
	issue_num := gojson.Json(str).Get("issue").Get("number").Tostring()
	c := api.NewClient(getToken)
	res := c.CreateGiteeIssueComment("Meta-OSS", "FenixscanX", issue_num, "hello")
	if res != nil {
		fmt.Println(res.Error())
		return
	}
}

func main() {
	http.HandleFunc("/", handlePostJson)
	http.ListenAndServe(":8008", nil)
}
