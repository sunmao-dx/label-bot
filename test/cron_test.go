package test

import (
	"encoding/csv"
	"fmt"
	gitee_utils "gitee.com/lizi/test-bot/src/gitee-utils"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCronRun(t *testing.T){
	//DoTime()
	//Remind()
	getUrl()
}

func DoByFixTime() {
	c := cron.New(cron.WithSeconds())
	//定时任务
	spec := "0 44 10 * * *" //cron表达式，每秒一次
	c.AddFunc(spec, func() {
		fmt.Println("11111")
	})
	c.Start()
	time.Sleep(time.Minute * 1)
}

func getToken() []byte {
	//return []byte("3250980ecd05028c40637004d97ce24a") // Clement Li
	return []byte("adb08695039522366c4a645e1e6a3dd4") // dx robot
}

func getUrl() {
	urlValues := url.Values{}
	urlValues.Add("labels","comp/data")
	resp, _ := http.PostForm("http://34.92.52.47/predict",urlValues)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	participants := string(body[:])
	fmt.Println(participants)
}

func Remind() {
	org := "mindspore"
	repo := "mindspore"
	c := gitee_utils.NewClient(getToken)

	fileName := "../src/data/IssueResAlert.csv"
	cntb, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	r2 := csv.NewReader(strings.NewReader(string(cntb)))
	ss, _ := r2.ReadAll()
	fmt.Println(ss)
	sz := len(ss)
	// 循环取数据
	for i := 0; i < sz; i++ {
		fmt.Println(ss[i])
		fmt.Println(ss[i][0]) //  key的数据  可以作为map的数据的值

		word := "hello, @" + ss[i][0] + " , Has this problem been resolved? If yes, please close this issue, thanks!" + "\n"
		wordCn := "你好, @" + ss[i][0] + " , 这个问题是否已经解决了呢？ 如果是的，请关闭这个issue， 谢谢！" + "\n"
		words := word + wordCn

		resRemind := c.CreateGiteeIssueComment(org, repo, ss[i][2], words)
		if resRemind != nil {
			fmt.Println(resRemind.Error())
			return
		}
	}
}

func DoTime() {
	owner := "mindspore"
	repo := "mindspore"
	state := "open"
	perPage := 100
	assigneeInit := ""
	cz := time.FixedZone("CST", 8*3600)
	fmt.Println(time.Now().AddDate(0,0,-60).In(cz).Format(time.RFC3339))
	//time.Sleep(time.Minute * 1)

	since := time.Now().AddDate(0,0,-15).In(cz)

	fromStr := time.Now().AddDate(0,0,-150).In(cz).Format(time.RFC3339)

	csvFile, err := os.Create("../src/data/IssueResAlert.csv")
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()
	csvFile.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(csvFile)
	c := gitee_utils.NewClient(getToken)

	_, response, res := c.ListIssues(owner, repo, state, fromStr, "", 1, perPage)
	if res != nil {
		fmt.Println(res.Error())
		return
	} else {
		if p, err := strconv.Atoi(response.Header.Get("total_page"));
		err == nil {
			for i := 1; i <= p; i++ {
				issues, _, res := c.ListIssues(owner, repo, state, fromStr, "", i, perPage)
				if res != nil {
					fmt.Println(res.Error())
					return
				} else {
					for _, issue := range issues {
						if issue.UpdatedAt.In(cz).Before(since) == false{
							continue
						}
						if issue.Assignee != nil {
							assigneeInit = issue.Assignee.Login
						}
						w.Write([]string{issue.User.Login, assigneeInit ,issue.Number, issue.UpdatedAt.In(cz).Format(time.RFC3339)})
						w.Flush()
						assigneeInit = ""
					}
				}
			}
		}
	}
}
