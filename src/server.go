package main

import (
	"fmt"
	"github.com/SmartsYoung/test/src/api"
	"github.com/widuu/gojson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func handlePostJson(writer http.ResponseWriter, request *http.Request) {
	h:= request.Header
	for key,value := range h{
		if key == "X-Gitee-Event"{
			str1 := strings.Join(value, "")
			if str1 == "Issue Hook"{
				issue_process(request)
			}
			if str1 == "Note Hook"{
				note_process(request)
			}

		}
	}
}

func getToken() []byte {
	return []byte("7f399efd043ee3d68e5edbf45326152b")
}


func note_process(request *http.Request) {
	s, _ := ioutil.ReadAll(request.Body) //把	body 内容读入字符串 s
	str := string(s)
	name := gojson.Json(str).Get("comment").Get("user").Get("name").Tostring()
	issue_num := gojson.Json(str).Get("issue").Get("number").Tostring()
	log.Println(name)
	log.Println(issue_num)



	//data := make(url.Values)
	if name != "test-bot"{
		//data["access_token"] = []string{"7f399efd043ee3d68e5edbf45326152b"}
		//data["body"] = []string{`["bug","kernel"]`}
		////把post表单发送给目标服务器
		////res, err := http.pa("https://gitee.com/api/v5/repos/Meta-OSS/FenixscanX/issues/"+issue_num, data)
		//if err != nil {
		//	fmt.Println(err.Error())
		//	return
		//}
		//defer res.Body.Close()
		//fmt.Println("post send success")
		c:= api.NewClient(getToken)
		res := c.AddIssueLabel("Meta-OSS", "FenixscanX", issue_num, "bug")
		if res.Error() != "" {
			fmt.Println(res.Error())
			return
		}
	}
}


func issue_process(request *http.Request) {
	s, _ := ioutil.ReadAll(request.Body) //把	body 内容读入字符串 s
	str := string(s)
	issue_num := gojson.Json(str).Get("issue").Get("number").Tostring()

	//这里添加post的body内容
	data := make(url.Values)
	data["access_token"] = []string{"7f399efd043ee3d68e5edbf45326152b"}
	data["body"] = []string{"Welcome to post an issue! Please offer me more issue labels. you can type /bug /help-wanted etc."}

	//把post表单发送给目标服务器
	res, err := http.PostForm("https://gitee.com/api/v5/repos/Meta-OSS/FenixscanX/issues/"+issue_num+"/comments", data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer res.Body.Close()
	fmt.Println("post send success")
}

func main () {
	http.HandleFunc("/", handlePostJson)
	http.ListenAndServe(":8008", nil)
}
