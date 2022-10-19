package projectreport

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/magiconair/properties"
	"github.com/perolo/confluence-client/client"
	"github.com/perolo/confluence-scripts/utilities"
	excelutils "github.com/perolo/excel-utils"
	"github.com/perolo/jira-client"
	"github.com/perolo/jira-scripts/jirautils"
)

type Header int

const (
	NAME Header = iota
	KEY
	CATEGORY
	PROJLEAD
	PROJSCHEME
	REQUIREMENT
	ACTION
	SUPPORTISSUE
	RISK
	LINK
	MONTHS3
	MONTHS1
	TIME
	SIZEHEADER
)

const allOk = "X"
const allFail = "-"

var headers = make([]string, SIZEHEADER)
var simplified = true

type aProject struct {
	value []string
}

func ProjectCategoryReport(propPtr string) {
	ProjectReportCategory(propPtr, "")
}

func ProjectReportCategory(propPtr, category string) {

	fmt.Printf("%%%%%%%%%%  ProjectReportCategory " + category + "%%%%%%%%%%%%%%\n")

	p := properties.MustLoadFile(propPtr, properties.ISO_8859_1)

	// or through Decode
	type Config struct {
		JiraHost   string `properties:"jirahost"`
		JiraUser   string `properties:"jirauser"`
		JiraPass   string `properties:"jirapass"`
		JiraToken  string `properties:"jiratoken"`
		ConfUser   string `properties:"confuser"`
		ConfPass   string `properties:"confpass"`
		ConfToken  string `properties:"conftoken"`
		UseToken   bool   `properties:"usetoken"`
		File       string `properties:"file"`
		Space      string `properties:"space"`
		Report     bool   `properties:"report"`
		ConfHost   string `properties:"confhost"`
		Attachment string `properties:"attachment"`
		Archivedwf string `properties:"archivedwf"`
	}
	var jiraClient *jira.Client
	var err error
	var cfg Config
	if err = p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}
	if cfg.UseToken {
		cfg.ConfPass = cfg.ConfToken
		cfg.JiraPass = cfg.JiraToken
	}
	headers[NAME] = "Project name"
	headers[KEY] = "Key"
	headers[CATEGORY] = "Category"
	headers[PROJLEAD] = "Project Lead"
	headers[PROJSCHEME] = "Permission Scheme"
	headers[REQUIREMENT] = "Requirements"
	headers[ACTION] = "Action"
	headers[SUPPORTISSUE] = "Support"
	headers[RISK] = "Risk"
	headers[LINK] = "Link"
	headers[TIME] = "Last Updated"
	allProjects := make(map[string]aProject)

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

	noProjects := 0
	noSkippedProjects := 0
	noEmpty := 0
	noNOCategory := 0

	projects, _, err := jiraClient.Project.GetList()
	excelutils.Check(err)
	for _, project := range *projects {
		if (category == "") || (project.ProjectCategory.Name == category) {
			var theProject aProject
			theProject.value = make([]string, SIZEHEADER)
			fmt.Printf("Project name: %s Key: %s\n", project.Name, project.Key)

			projPerm, closedDown := jirautils.GetPermissionScheme(jiraClient, project.Key, cfg.Archivedwf)

			if simplified && closedDown {
				fmt.Printf("Project skipped : %s \n", project.Name)
				noSkippedProjects++
			} else {

				// Issues updated in last 3 Months
				jql := "project=\"" + project.Key + "\" AND updated >= -90d AND updated <= \"0\""
				sres, _, jerr := jiraClient.Issue.Search(jql, &jira.SearchOptions{StartAt: 0, MaxResults: 1})
				count90d := ""
				if jerr != nil {
					count90d = "err"
				}
				if sres == nil {
					panic(sres)
				} else {
					count90d = fmt.Sprintf("%v", sres.Total)
				}

				// Issues updated in last Month
				jql = "project=\"" + project.Key + "\" AND updated >= -30d AND updated <= \"0\""
				sres, _, jerr = jiraClient.Issue.Search(jql, &jira.SearchOptions{StartAt: 0, MaxResults: 1})
				count30d := ""
				if jerr != nil {
					count30d = "err"
				}
				if sres == nil {
					panic(sres)
				} else {
					count30d = fmt.Sprintf("%v", sres.Total)
				}

				// Updated Date
				jql = "project=\"" + project.Key + "\" ORDER BY updated DESC"
				sres, _, jerr = jiraClient.Issue.Search(jql, &jira.SearchOptions{StartAt: 0, MaxResults: 1})
				ticketTime := ""
				if jerr != nil {
					ticketTime = "err"
				}
				if sres == nil {
					panic(sres)
				}
				if len(sres.Issues) == 1 {
					ticketTime = ((time.Time)(sres.Issues[0].Fields.Updated)).Format("2006-01-02 15:04:05")
				} else {
					ticketTime = ticketTime + " empty"
					noEmpty++
				}
				var detailProj *jira.Project
				for i := 0; ; i++ {
					detailProj, _, err = jiraClient.Project.Get(project.Key)
					if err == nil {
						break
					}
					time.Sleep(2 * time.Second)
				}
				excelutils.Check(err)

				theProject.value[REQUIREMENT] = allFail
				theProject.value[ACTION] = allFail
				theProject.value[SUPPORTISSUE] = allFail
				theProject.value[RISK] = allFail
				for _, issType := range detailProj.IssueTypes {
					switch issType.Name {
					case "Requirement":
						theProject.value[REQUIREMENT] = allOk
					case "Action":
						theProject.value[ACTION] = allOk
					case "Support Issue":
						theProject.value[SUPPORTISSUE] = allOk
					case "Risk":
						theProject.value[RISK] = allOk

					}
				}
				// priorityscheme
				// notificationscheme

				theProject.value[NAME] = project.Name
				theProject.value[KEY] = project.Key
				theProject.value[CATEGORY] = project.ProjectCategory.Name
				if project.ProjectCategory.Name == "" {
					noNOCategory++
				}
				theProject.value[PROJLEAD] = detailProj.Lead.DisplayName
				theProject.value[PROJSCHEME] = projPerm
				theProject.value[LINK] = cfg.JiraHost + "/projects/" + project.Key + "/summary/statistics"
				theProject.value[MONTHS1] = count30d
				theProject.value[MONTHS3] = count90d
				theProject.value[TIME] = ticketTime
				allProjects[project.Name] = theProject

			}
			noProjects++
		}
	}
	fmt.Printf("Projects : %d \n", noProjects)
	excelutils.Check(err)

	excelFile := ""
	var copt client.OperationOptions
	if category == "" {
		copt.Title = "Project Report"
		cfg.Attachment = "ProjectCategoryReport" + ".xlsx"
		excelFile = fmt.Sprintf(cfg.File, "-all")
	} else {
		copt.Title = "Project Report"
		cfg.Attachment = "ProjectReport_" + category + ".xlsx"
		excelFile = fmt.Sprintf(cfg.File, "-"+category)
	}
	excelutils.Check(err)
	excelutils.NewFile()

	excelutils.SetCellFontHeader()
	excelutils.WiteCellln("Introduction")

	excelutils.WiteCellln("Please Do not edit this page!")
	excelutils.WiteCellln("This page is created by the projectreport script: github.com/perolo/jira-scripts/")
	t := time.Now()

	excelutils.WiteCellln("Created by: " + cfg.ConfUser + " : " + t.Format(time.RFC3339))
	excelutils.WiteCellln("")

	if simplified {
		excelutils.WiteCellln("Projects with Closed Down or Archived Permission Schemes have been removed from report\n")
		excelutils.WiteCellln("")
	}

	excelutils.WiteCellln("Projects with \"empty\" Last Updated are completely empty without issues. \n")
	nostr := fmt.Sprintf("Projects : %d \n", noProjects)
	excelutils.WiteCellln(nostr)
	nostr = fmt.Sprintf("No Skipped Projects : %d \n", noSkippedProjects)
	excelutils.WiteCellln(nostr)
	nostr = fmt.Sprintf("Empty Projects : %d \n", noEmpty)
	excelutils.WiteCellln(nostr)
	nostr = fmt.Sprintf("Projects without Category: %d \n", noNOCategory)
	excelutils.WiteCellln(nostr)

	excelutils.WiteCellln("")

	excelutils.SetCellFontHeader2()
	excelutils.WiteCellln("JIRA Projects")

	excelutils.AutoFilterStart()
	excelutils.WriteColumnsHeaderRotln(headers)

	var sortednames []string
	for k := range allProjects {
		sortednames = append(sortednames, k)
	}
	sort.Strings(sortednames)
	for _, spacename := range sortednames {
		excelutils.WriteColumnsln(allProjects[spacename].value)
	}
	excelutils.SetAutoColWidth()
	excelutils.AutoFilterEnd()
	excelutils.SetColWidth("A", "A", 60)

	excelutils.SaveAs(excelFile)

	if cfg.Report {
		var config = client.ConfluenceConfig{}
		config.Username = cfg.ConfUser
		config.Password = cfg.ConfPass
		config.UseToken = cfg.UseToken
		config.URL = cfg.ConfHost
		config.Debug = false
		confluenceClient := client.Client(&config)
		copt.SpaceKey = cfg.Space
		copt.BodyOnly = true

		_, name := filepath.Split(cfg.Attachment)
		err = utilities.AddAttachmentAndUpload(confluenceClient, copt, name, copt.Filepath, "Created by Project Report")
		if err != nil {
			panic(err)
		}
	}
}
