package main

import (
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/jira-client"
	"log"
	"strings"
	"time"
	"github.com/trivago/tgo/tcontainer"
)

func main() {

	propPtr := flag.String("prop", "git.aa.st/perolo/jira-utils/test-jira.properties", "a string")

	flag.Parse()

	p := properties.MustLoadFile(*propPtr, properties.ISO_8859_1)

	// or through Decode
	type Config struct {
		JiraHost string `properties:"jirahost"`
		User string `properties:"user"`
		Pass string `properties:"password"`
		DestinationProject string `properties:"destinationproject"`
		JQL string `properties:"jql"`
		CloneSubtasks bool `properties:"clonesubtasks"`

	}
	var cfg Config
	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(cfg.User),
		Password: strings.TrimSpace(cfg.Pass),
	}

	jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(cfg.JiraHost))
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}
	//Get all source issues
	var allIssues [] jira.Issue
	cont :=true
	start:=0
	max := 50
	for cont {
		//qjql := url.QueryEscape(propConfig.JQL)
		sres, _, err := jiraClient.Issue.Search(cfg.JQL, &jira.SearchOptions{StartAt:start, MaxResults:max})
		if err != nil {
			panic(err)
		}
		allIssues = append(allIssues, sres.Issues...)
		if len(sres.Issues) !=max {
			cont=false
		} else {
			start = start + max
		}
	}

	var sourceIssueTypes  map[string]string
	sourceIssueTypes = make(map[string]string,10)

	//Get source issue types
	for _, iss := range allIssues {
		if _, ok :=sourceIssueTypes[iss.Fields.Type.Name]; !ok{
			sourceIssueTypes[iss.Fields.Type.Name] = iss.Fields.Type.Name
		}
	}

	//Check source issue types exist in destination
	dproj,_, err := jiraClient.Project.Get(cfg.DestinationProject)

	for _, siss := range sourceIssueTypes {
		found := false
		for _, diss := range dproj.IssueTypes {
			if diss.Name == siss {
				found = true
			}
		}
		if siss == "Sub-task"  {
			fmt.Printf("Do not include subtasks in JQL, handled with option : clonesubtasks \n")
			panic(err)
		}
		if !found {
			fmt.Printf("All IssueTypes must be available in destination Project : %s\n", dproj.Name)
			panic(err)
		}
	}

	//Loop all issues
	for _, iss := range allIssues {

		createdIss, err := createIssue(iss, dproj, jiraClient)
		if err != nil {
			fmt.Printf("Result: %s\n", err.Error())
			panic(err)
		}

		if cfg.CloneSubtasks {
			for _, sisssub := range iss.Fields.Subtasks {
				_, err := createSubIssue(*sisssub, dproj, jiraClient, *createdIss)
				if err != nil {
					fmt.Printf("Result: %s\n", err.Error())
					panic(err)
				}

			}

		}
		aLink := new(jira.IssueLink)
		aLink.InwardIssue = new(jira.Issue)
		aLink.OutwardIssue = new(jira.Issue)
		aLink.InwardIssue.Key = iss.Key
		aLink.Type.Name = "Relates"
		aLink.OutwardIssue.Key = createdIss.Key

		_, err = jiraClient.Issue.AddLink(aLink)
		if err != nil {
			fmt.Printf("Result: %s\n", err.Error())
			panic(err)
		}

		aComment := new(jira.Comment)
		aComment.Body = "Issue copied using jiraclonescript by " + cfg.User + " " + time.Now().Format("2006-01-02 15:04:05")
		_, _, err = jiraClient.Issue.AddComment(createdIss.Key, aComment)
		if err != nil {
			fmt.Printf("Result: %s\n", err.Error())
			panic(err)
		}
	}
	fmt.Printf("Created : %v Issues\n", len(allIssues))
}

func createIssue(iss jira.Issue, dproj *jira.Project, jiraClient *jira.Client) (*jira.Issue, error) {
	//Clone, link, Label and Comment issues
	newIssue := new(jira.Issue)
	issueFields := new(jira.IssueFields)
	newIssue.Fields = issueFields

	newIssue.Fields.Summary = iss.Fields.Summary
	newIssue.Fields.Description = iss.Fields.Description
	newIssue.Fields.Project.Key = dproj.Key
	newIssue.Fields.Type.Name = iss.Fields.Type.Name
	if iss.Fields.Type.Name == "Epic" {
		newIssue.Fields.Summary = iss.Fields.Summary + " - Cloned"
		newIssue.Fields.Unknowns = tcontainer.NewMarshalMap()
		newIssue.Fields.Unknowns["customfield_10017"] = iss.Fields.Summary + " - Cloned"
	}
	newIssue.Fields.Labels = iss.Fields.Labels

	createdIss, _, err := jiraClient.Issue.Create(newIssue)

	if err != nil {
		fmt.Printf("Result: %s\n", err.Error())
		panic(err)
	}
	fmt.Printf("    Created Issue: %s \n", createdIss.Key)
	return createdIss, err
}
func createSubIssue(iss jira.Subtasks, dproj *jira.Project, jiraClient *jira.Client, parent jira.Issue) (*jira.Issue, error) {
	//Clone, link, Label and Comment issues
	newIssue := new(jira.Issue)
	issueFields := new(jira.IssueFields)
	newIssue.Fields = issueFields

	newIssue.Fields.Summary = iss.Fields.Summary
	newIssue.Fields.Description = iss.Fields.Description
	newIssue.Fields.Project.Key = dproj.Key
	newIssue.Fields.Type.Name = iss.Fields.Type.Name


	newIssue.Fields.Parent = &jira.Parent{ID:parent.ID,Key:parent.Key}

	newIssue.Fields.Labels = iss.Fields.Labels

	createdIss, _, err := jiraClient.Issue.Create(newIssue)

	if err != nil {
		fmt.Printf("Result: %s\n", err.Error())
		panic(err)
	}
	fmt.Printf("    Created Issue: %s \n", createdIss.Key)
	return createdIss, err
}
