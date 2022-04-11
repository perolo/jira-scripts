package customfieldreport

import (
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/excel-utils"
	"github.com/perolo/jira-client"
	"github.com/perolo/jira-scripts/jirautils"
	"log"
	"strings"
	"time"
)

var headers []string

type customField struct {
	reportType string
	project    string
	//	id              string
	projCategory    string
	projPerm        string
	archivedProject bool
	issueCount      int
	issueUpdateDate time.Time
	projCount       int
	screensCount    int
	customField     jira.CustomFieldsType
}

type Config struct {
	JiraHost     string `properties:"jirahost"`
	ConfHost     string `properties:"confhost"`
	JiraUser     string `properties:"jirauser"`
	ConfUser     string `properties:"confuser"`
	ConfPass     string `properties:"confpass"`
	JiraPass     string `properties:"jirapass"`
	JiraToken    string `properties:"jiratoken"`
	ConfToken    string `properties:"conftoken"`
	UseToken     bool   `properties:"usetoken"`
	Upload       bool   `properties:"upload"`
	Space        string `properties:"space"`
	File         string `properties:"file"`
	DeleteUnused bool   `properties:"deleteunused"`
	//	Attachment string `properties:"attachment"`
	Archivedwf string `properties:"archivedwf"`
	//		Bindusername string `properties:"bindusername"`
	//		Bindpassword string `properties:"bindpassword"`
	//		BaseDN           string `properties:"basedn"`
}

func CustomFieldReport(propPtr string) { //nolint:funlen, gocyclo

	var jiraClient *jira.Client
	var err error
	fmt.Printf("%%%%%%%%%%  JIRA Custom Field Report %%%%%%%%%%%%%%\n")

	p := properties.MustLoadFile(propPtr, properties.ISO_8859_1)

	var cfg Config
	if err2 := p.Decode(&cfg); err2 != nil {
		log.Fatal(err2)
	}
	// Temporary workaround solution - need to find better?

	allCustomFields := make(map[string]customField)
	projectCustomFields := make(map[string]customField)
	projectUsageFields := make(map[string]customField)
	/*
		var config = client.ConfluenceConfig{}
		config.Username = cfg.ConfUser
		config.Password = cfg.ConfPass
		config.UseToken = cfg.UseToken
		config.URL = cfg.ConfHost

		confluence := client.Client(&config)
	*/
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

	//jiraClient.Debug = true
	cont := true
	start := 0
	max := 1
	limit := 0 // Used for debugging - only get a subset
	for cont {
		projloop := 0
		// Only searching 1 customfield at a time is a confirmed Jira bug  -
		fields, _, err2 := jiraClient.Field.GetAllCustomFields(&jira.FieldOptions{StartAt: start, MaxResults: max})
		jirautils.Check(err2)
		for _, field := range fields.Values {
			//limit++
			fmt.Printf("CustomField: %s \n", field.Name)
			var aField customField
			aField.customField = field
			if field.Name == "OTHER" {
				fmt.Printf("CustomField: %s  \n", field.Name)
			}

			// Need to investigate these custom fields - not sure how to solve this problem in a more generic way?
			// Throws an exception - not possible to compare with EMPTY
			if field.IsAllProjects && field.Name != "Development" && field.Name != "Progress" && field.Name != "Rank" && field.Name != "Zephyr Teststep" {
				//fmt.Printf("CustomField: %s\n", field.Name)
				aField.reportType = "Custom Field Global"

				aField.screensCount = field.ScreensCount
				jql := fmt.Sprintf("cf[%v] is not EMPTY ", field.NumericID)
				issues, _, err3 := jiraClient.Issue.Search(jql, &jira.SearchOptions{StartAt: 0, MaxResults: 1})
				if err3 != nil {
					break
					//panic(err)
				}
				aField.issueCount = issues.Total
				if issues.Total > 0 {
					aField.issueUpdateDate = (time.Time)(issues.Issues[0].Fields.Updated)
				} else {
					if !field.IsLocked && field.Name != "Story Points" {
						if cfg.DeleteUnused {
							if jirautils.QueryUser(fmt.Sprintf("Remove CustomField?: %s [y/n]: ", field.Name)) {
								//Not available in Server
								_, _, err4 := jiraClient.Field.DeleteCustomField(field.ID)
								if err4 != nil {
									panic(err4)
								}

							}
						}
					}
				}
				if issues.Total > 0 {
					projloop = 1
					remaining := issues.Total
					var projects, lastproj string
					lastproj = issues.Issues[0].Fields.Project.Name
					projects = "\"" + issues.Issues[0].Fields.Project.Name + "\""
					for remaining > 0 {
						// Should be possible to optimize...
						fmt.Printf("CustomField: %s Project: %s \n", field.Name, lastproj)
						jql2 := fmt.Sprintf("cf[%v] is not EMPTY AND project in (\"%s\") ORDER BY updatedDate DESC", field.NumericID, lastproj)
						issues2, _, err4 := jiraClient.Issue.Search(jql2, &jira.SearchOptions{StartAt: 0, MaxResults: 1})
						if err4 != nil {
							panic(err4)
						}
						if (issues2.Total) == 0 {
							// Something strange
							fmt.Printf("Something strange: \n")
						} else {
							var bField customField
							bField.reportType = "Custom Field Usage"
							bField.customField = field
							bField.project = lastproj
							bField.issueCount = issues2.Total
							bField.issueUpdateDate = (time.Time)(issues2.Issues[0].Fields.Updated)
							projectUsageFields[(lastproj + field.Name)] = bField

						}
						jql3 := fmt.Sprintf("cf[%v] is not EMPTY AND project not in (%s)", field.NumericID, projects)
						issues3, _, err2 := jiraClient.Issue.Search(jql3, &jira.SearchOptions{StartAt: 0, MaxResults: 1})
						if err2 != nil {
							panic(err2)
						}
						if (issues3.Total) == 0 { //the last
							//fmt.Printf("The Last: \n")
							remaining = 0
						} else { //still some
							lastproj = issues3.Issues[0].Fields.Project.Name
							projects = projects + ",\"" + lastproj + "\""
							remaining -= issues3.Total
							projloop++
						}
					}
				}
				aField.projCount = projloop
			} else {
				jql := fmt.Sprintf("cf[%v] is not EMPTY ", field.NumericID)
				issues, _, err2 := jiraClient.Issue.Search(jql, &jira.SearchOptions{StartAt: 0, MaxResults: 1})
				if err2 != nil {
					break
					//panic(err)
				}
				aField.issueCount = issues.Total
				if issues.Total > 0 {
					aField.issueUpdateDate = (time.Time)(issues.Issues[0].Fields.Updated)
				} else {
					if !field.IsLocked {
						if cfg.DeleteUnused {
							if jirautils.QueryUser(fmt.Sprintf("Remove CustomField?: %s [y/n]: ", field.Name)) {
								//Not available in Server
								_, _, err3 := jiraClient.Field.DeleteCustomField(field.ID)
								if err3 != nil {
									panic(err3)
								}

							}
						}
					}
				}
				aField.projCount = field.ProjectsCount
				aField.screensCount = field.ScreensCount
				aField.reportType = "Custom Field Project"
			}
			allCustomFields[field.Name] = aField
			if limit > 100 {
				cont = false
			}
		}
		start = start + len(fields.Values)
		if start > fields.Total {
			cont = false
		}
		if limit > 100 {
			cont = false
		}
	}
	limit = 0
	projects, _, err := jiraClient.Project.GetList()
	if err != nil {
		fmt.Printf("Result: %v\n", err.Error())
		panic(err)
	}
	for _, project := range *projects {
		fmt.Printf("Project: %s\n", project.Name)
		fields, _, err2 := jiraClient.Field.GetAllCustomFields(&jira.FieldOptions{StartAt: 0, MaxResults: 5, ProjectIds: project.ID})
		if err2 != nil {
			fmt.Printf("Result: %v\n", err2.Error())
			panic(err2)
		}
		projPerm, archived := jirautils.GetPermissionScheme(jiraClient, project.Key, cfg.Archivedwf)

		var cField customField
		cField.reportType = "Project"
		cField.project = project.Name
		cField.projCategory = project.ProjectCategory.Name
		cField.projPerm = projPerm
		cField.archivedProject = archived
		cField.projCount = len(fields.Values)
		projectCustomFields[project.Name] = cField

		for _, field := range fields.Values {

			var aField customField
			aField.customField = field
			aField.reportType = "Project Use"
			aField.project = project.Name
			aField.projCount = len(fields.Values)
			aField.projCategory = project.ProjectCategory.Name
			aField.projPerm = projPerm
			cField.archivedProject = archived
			projectCustomFields[field.Name] = aField
		}
		//limit++
		if limit > 100 {
			break
		}

	}

	file := fmt.Sprintf(cfg.File, "-Jira-CustomFieds")
	//	defer os.Remove(f.Name())

	excelutils.NewFile()
	excelutils.SetCellFontHeader()
	excelutils.WiteCellln("Introduction")
	excelutils.WiteCellln("Please Do not edit this page!")
	excelutils.WiteCellln("This page is created by the User Report script: " + "https://git.aa.st/perolo/jira-utils/CustomfieldReport")
	t := time.Now()

	excelutils.WiteCellln("Created by: " + cfg.ConfUser + " : " + t.Format(time.RFC3339))
	excelutils.WiteCellln("")

	excelutils.SetCellFontHeader2()
	excelutils.WiteCellln("Custom Fields")

	headers = append(headers, "Report")
	headers = append(headers, "Custom Field")
	headers = append(headers, "ID")
	headers = append(headers, "Global")
	headers = append(headers, "Customfield Projects")
	headers = append(headers, "Customfield Screen")
	headers = append(headers, "Issue Count")
	headers = append(headers, "Issue Update")
	headers = append(headers, "Projects Count")
	headers = append(headers, "Screens Count")
	headers = append(headers, "Project")
	headers = append(headers, "Category")
	headers = append(headers, "Archived")
	headers = append(headers, "Permission Scheme")
	headers = append(headers, "Description")

	excelutils.AutoFilterStart()
	excelutils.SetRowHeight(20)
	excelutils.WriteColumnsHeaderln(headers)

	for _, field := range allCustomFields {
		writeLine(field)
	}
	for _, field := range projectCustomFields {
		writeLine(field)
	}
	for _, field := range projectUsageFields {
		writeLine(field)
	}
	excelutils.SetAutoColWidth()
	excelutils.AutoFilterEnd()
	excelutils.SetColWidth("A", "A", 40)

	// Save xlsx file by the given path.
	excelutils.SaveAs(file)

	/*
		if cfg.Upload {
				var copt client.OperationOptions
				copt.Title = "Jira Custom Field Report"
				copt.SpaceKey = cfg.Space
				copt.BodyOnly = true
				_, name := filepath.Split(file)
				utilities.CheckPageExists(copt, confluence)
				err = utilities.AddAttachmentAndUpload(confluence, copt, name, file, "Created by Custom Field Report")
				if err != nil {
					panic(err)
				}
		}
	*/
}

func writeLine(field customField) {
	excelutils.WiteCellnc(field.reportType)
	excelutils.WiteCellnc(field.customField.Name)
	excelutils.WiteCellnc(field.customField.ID)
	excelutils.WiteCellnc(fmt.Sprintf("%t", field.customField.IsAllProjects))
	if field.customField.ProjectsCount != 0 {
		excelutils.WiteCellnc(fmt.Sprintf("%v", field.customField.ProjectsCount))
	} else {
		excelutils.WiteCellnc("") //ProjectsCount
	}
	if field.customField.ScreensCount != 0 {
		excelutils.WiteCellnc(fmt.Sprintf("%v", field.customField.ScreensCount))
	} else {
		excelutils.WiteCellnc("") //ProjectsCount
	}
	excelutils.WiteCellnc(fmt.Sprintf("%v", field.issueCount))
	if time.Time.IsZero(field.issueUpdateDate) {
		excelutils.WiteCellnc("")
	} else {
		excelutils.WiteCellnc((field.issueUpdateDate).Format("2006-01-02"))
	}
	if field.projCount != 0 {
		excelutils.WiteCellnc(fmt.Sprintf("%v", field.projCount))
	} else {
		excelutils.WiteCellnc("") // projCount
	}
	excelutils.WiteCellnc(field.project)
	excelutils.WiteCellnc(field.projCategory)
	excelutils.WiteCellnc(fmt.Sprintf("%t", field.archivedProject))
	excelutils.WiteCellnc(field.projPerm)
	excelutils.WiteCellnc(field.customField.Description)
	excelutils.NextLine()
}
