package cloneissues

import (
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/jira-client"
	"github.com/trivago/tgo/tcontainer"
	"log"
	"reflect"
	"strings"
	"time"
)

type Config struct {
	JiraHost           string `properties:"jirahost"`
	JiraUser           string `properties:"jirauser"`
	UseToken           bool   `properties:"usetoken"`
	JiraPass           string `properties:"jirapass"`
	JiraToken          string `properties:"jiratoken"`
	DestinationProject string `properties:"destinationproject"`
	JQL                string `properties:"jql"`
	CloneSubtasks      bool   `properties:"clonesubtasks"`
	AddLabel           string `properties:"label"`
	Epic               string `properties:"epic"`
}

func CloneIssues(propPtr string) { //nolint:funlen
	var jiraClient *jira.Client
	var err error

	p := properties.MustLoadFile(propPtr, properties.ISO_8859_1)

	var cfg Config
	if err = p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	if cfg.UseToken {
		tp := jira.BearerAuthTransport{
			Token: strings.TrimSpace(cfg.JiraToken),
		}
		jiraClient, err = jira.NewClient(tp.Client(), strings.TrimSpace(cfg.JiraHost))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		tp := jira.BasicAuthTransport{
			Username: strings.TrimSpace(cfg.JiraUser),
			Password: strings.TrimSpace(cfg.JiraPass),
		}
		jiraClient, err = jira.NewClient(tp.Client(), strings.TrimSpace(cfg.JiraHost))
		if err != nil {
			log.Fatal(err)
		}
	}
	jiraClient.Debug = true

	//Get all source issues
	var allIssues []jira.Issue
	cont := true
	start := 0
	max := 50
	for cont {
		//qjql := url.QueryEscape(propConfig.JQL)
		sres, _, err2 := jiraClient.Issue.Search(cfg.JQL, &jira.SearchOptions{StartAt: start, MaxResults: max})
		if err2 != nil {
			panic(err2)
		}
		allIssues = append(allIssues, sres.Issues...)
		if len(sres.Issues) != max {
			cont = false
		} else {
			start = start + max
		}
	}
	var sourceIssueTypes = make(map[string]string, 100) //TODO What if?
	//Get source issue types
	for _, iss := range allIssues {
		if _, ok := sourceIssueTypes[iss.Fields.Type.Name]; !ok {
			sourceIssueTypes[iss.Fields.Type.Name] = iss.Fields.Type.Name
		}
	}
	//Check source issue types exist in destination
	dproj, _, err := jiraClient.Project.Get(cfg.DestinationProject)
	for _, siss := range sourceIssueTypes {
		found := false
		for _, diss := range dproj.IssueTypes {
			if diss.Name == siss {
				found = true
			}
		}
		if siss == "Sub-task" {
			fmt.Printf("Do not include subtasks in JQL, handled with option : clonesubtasks \n")
			panic(err)
		}
		if !found {
			fmt.Printf("All IssueTypes must be available in destination Project : %s\n", dproj.Name)
			panic(err)
		}
	}
	issuecounter := 0
	//Loop all issues
	for _, iss := range allIssues {
		createdIss, err2 := createIssue(iss, dproj, jiraClient, cfg)
		if err2 != nil {
			fmt.Printf("Result: %s\n", err2.Error())
			panic(err2)
		}
		issuecounter++
		if cfg.CloneSubtasks {
			for _, sisssub := range iss.Fields.Subtasks {
				issuecounter++
				var thesub *jira.Issue
				// fix to get around subtasks to not contain all info
				thesub, _, err2 = jiraClient.Issue.Get(sisssub.ID, nil)
				if err2 != nil {
					fmt.Printf("Result: %s\n", err2.Error())
					panic(err2)
				}
				_, err3 := createSubIssue(*thesub, dproj, jiraClient, *createdIss, cfg)
				if err3 != nil {
					fmt.Printf("Result: %s\n", err3.Error())
					panic(err3)
				}
			}
		}
		aLink := new(jira.IssueLink)
		aLink.InwardIssue = new(jira.Issue)
		aLink.OutwardIssue = new(jira.Issue)
		aLink.InwardIssue.Key = iss.Key
		aLink.Type.Name = "Relates"
		aLink.OutwardIssue.Key = createdIss.Key
		_, err2 = jiraClient.Issue.AddLink(aLink)
		if err2 != nil {
			fmt.Printf("Result: %s\n", err2.Error())
			panic(err2)
		}
		aComment := new(jira.Comment)
		aComment.Body = "Issue copied using jiraclonescript by " + cfg.JiraUser + " " + time.Now().Format("2006-01-02 15:04:05")
		_, _, err2 = jiraClient.Issue.AddComment(createdIss.Key, aComment)
		if err2 != nil {
			fmt.Printf("Result: %s\n", err2.Error())
			panic(err2)
		}
	}
	fmt.Printf("Created : %v Issues\n", issuecounter)
}

func createIssue(iss jira.Issue, dproj *jira.Project, jiraClient *jira.Client, theCfg Config) (*jira.Issue, error) {
	//Clone, link, SpaceCategory and Comment issues
	newIssue := new(jira.Issue)
	issueFields := new(jira.IssueFields)
	newIssue.Fields = issueFields
	newIssue.Fields.Project.Key = dproj.Key

	emeta, _, err := jiraClient.Issue.GetEditMeta(&iss)
	if err == nil {
		//fmt.Printf("Emeta: %s\n", emeta)
		for key, value := range emeta.Fields {
			var f = value.(map[string]interface{})
			req := reflect.ValueOf(f["required"]).Bool()
			//fmt.Printf("Key: %s required: %t \n", key, req)
			switch key {
			case "issuetype":
				newIssue.Fields.Type.Name = iss.Fields.Type.Name
			case "summary":
				newIssue.Fields.Summary = iss.Fields.Summary
			case "description":
				newIssue.Fields.Description = iss.Fields.Description
			case "components":
				{
					if len(iss.Fields.Components) != 0 {
						fmt.Printf("Components are not cloned! \n")
					}
				}
			case "labels":
				{
					if len(theCfg.AddLabel) == 0 {
						newIssue.Fields.Labels = iss.Fields.Labels
					} else {
						newIssue.Fields.Labels = append(iss.Fields.Labels, theCfg.AddLabel)
					}
				}
			case "reporter":
				newIssue.Fields.Reporter = iss.Fields.Reporter
			case theCfg.Epic:
				{
					if iss.Fields.Type.Name == "Epic" {
						newIssue.Fields.Summary = iss.Fields.Summary + " - Cloned"
						newIssue.Fields.Unknowns = tcontainer.NewMarshalMap()
						newIssue.Fields.Unknowns[theCfg.Epic] = iss.Fields.Summary + " - Cloned"
					}
				}
			default:
				if req {
					fmt.Printf("Unhandeled Mandatory Key: %s required: %t \n", key, req)
					panic(err)
				}

			}
		}
	}

	createdIss, _, err := jiraClient.Issue.Create(newIssue)

	if err != nil {
		fmt.Printf("Result: %s\n", err.Error())
		panic(err)
	}
	fmt.Printf("    Created Issue: %s \n", createdIss.Key)
	return createdIss, err
}
func createSubIssue(iss jira.Issue, dproj *jira.Project, jiraClient *jira.Client, parent jira.Issue, theCfg Config) (*jira.Issue, error) {
	//Clone, link, SpaceCategory and Comment issues
	newIssue := new(jira.Issue)
	issueFields := new(jira.IssueFields)
	newIssue.Fields = issueFields

	newIssue.Fields.Summary = iss.Fields.Summary
	newIssue.Fields.Description = iss.Fields.Description
	newIssue.Fields.Project.Key = dproj.Key
	newIssue.Fields.Type.Name = iss.Fields.Type.Name

	newIssue.Fields.Parent = &jira.Parent{ID: parent.ID, Key: parent.Key}

	if len(theCfg.AddLabel) == 0 {
		newIssue.Fields.Labels = iss.Fields.Labels
	} else {
		newIssue.Fields.Labels = append(iss.Fields.Labels, theCfg.AddLabel)
	}

	createdIss, _, err := jiraClient.Issue.Create(newIssue)

	if err != nil {
		fmt.Printf("Result: %s\n", err.Error())
		panic(err)
	}
	fmt.Printf("    Created Issue: %s \n", createdIss.Key)
	return createdIss, err
}
