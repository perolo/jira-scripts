package jirajqljson

import (
	"encoding/json"
	"fmt"
	"git.aa.st/perolo/confluence-utils/Utilities/htmlutils"
	"github.com/atotto/clipboard"
	"github.com/magiconair/properties"
	"github.com/perolo/jira-client"
	"log"
	"os"
	"strings"
)

// or through Decode
type Config struct {
	Host string `properties:"host"`
	User string `properties:"user"`
	Pass string `properties:"password"`
}
type Data struct {
	Key string `json:"key"`
	Summary string `json:"summary"`
	Assignee string `json:"assignee"`
	Link string `json:"link"`
	StatusCategory string `json:"statuscategory"`
}

var cfg Config

func JiraJQLjson(propPtr string) {

	fmt.Printf("%%%%%%%%%%  jirajqljson %%%%%%%%%%%%%%\n")

	text, _ := clipboard.ReadAll()
	fmt.Println(text)

	p := properties.MustLoadFile(propPtr, properties.ISO_8859_1)

	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(cfg.User),
		Password: strings.TrimSpace(cfg.Pass),
	}

	jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(cfg.Host))
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}
	var jql string

//	theFile := "data.json"

	jql = "project=\"Scrum Test Project\" AND issuetype =Epic"

	var data []Data
	cont := true
	start := 0
	max := 50
	for cont {
		//qjql := url.QueryEscape(propConfig.JQL)
		issues, _, err := jiraClient.Issue.Search(jql, &jira.SearchOptions{StartAt: start, MaxResults: max})
		if err != nil {
			panic(err)
		}
		for _, iss := range issues.Issues {
			fmt.Printf("Issue: %s \n", iss.Key)
			data = append(data, Data{Key: iss.Key, Summary: iss.Fields.Summary,Assignee: iss.Fields.Assignee.Name,Link: iss.Self, StatusCategory: iss.Fields.Status.StatusCategory.Name})
		}
		if len(issues.Issues) != max {
			cont = false
		} else {
			start = start + max
		}

	}

	buf, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	//	f, err := ioutil.TempFile(os.TempDir(), "data*.json")
	f, err := os.Create("C://temp/data.json")
	htmlutils.Check(err)
	_, err = f.Write(buf)
	htmlutils.Check(err)
	err = f.Close()
	htmlutils.Check(err)
	fmt.Printf("File created OK\n")

}


