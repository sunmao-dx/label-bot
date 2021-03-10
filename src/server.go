package main

import (
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
	return []byte("7f399efd043ee3d68e5edbf45326152b")
}

func getTokenstr() string {
	return "7f399efd043ee3d68e5edbf45326152b"
}

func note_process(request *http.Request) {
	token := getTokenstr()
	s, _ := ioutil.ReadAll(request.Body) //把	body 内容读入字符串 s
	str := string(s)
	name := gojson.Json(str).Get("comment").Get("user").Get("name").Tostring()
	issue_num := gojson.Json(str).Get("issue").Get("number").Tostring()

	if name != "test-bot" {
		c := api.NewClient(getToken)
		res := c.AddIssueLabel("Meta-OSS", "FenixscanX", issue_num, "kernel")
		if res != nil {
			fmt.Println(res.Error())
			return
		}
		c1 := api.NewClient(getToken)
		res = c1.AddIssueAssignee("Meta-OSS", "FenixscanX", issue_num, token, "clement_li")
		if res != nil {
			fmt.Println(res.Error())
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
