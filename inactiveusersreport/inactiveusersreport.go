package inactiveusersreport

import (
	"context"
	"flag"
	"fmt"
	"git.aa.st/perolo/confluence-utils/Utilities"
	"github.com/magiconair/properties"
	"github.com/perolo/confluence-prop/client"
	"github.com/perolo/confluence-scripts/utilities"
	excelutils "github.com/perolo/excel-utils"
	"github.com/perolo/jira-client"
	"log"
	"path/filepath"
	"strings"
	"time"
)

type ReportConfig struct {
	Host            string `properties:"host"`
	ConfHost        string `properties:"confhost"`
	User            string `properties:"user"`
	Pass            string `properties:"password"`
	ProjectCategory string `properties:"projectcategory"`
	File            string `properties:"file"`
	Simple          bool   `properties:"simple"`
	Report          bool   `properties:"report"`
	//	RolesReport      bool   `properties:"rolesreport"`
	//	ExpandGroups     bool   `properties:"expandgroups"`
	//	PermissionReport bool   `properties:"permissionreport"`
}

type ProjectUserType struct {
	projectName      string
	permissionScheme string
	user             string
	group            string
	role             string
	displayName      string
	adminPermission  bool
	teamPermission   bool
	browsePermission bool
	active           bool
}

var allProjectUsers map[string]ProjectUserType

func InactiveUserReport(propPtr string) {

	flag.Parse()

	p := properties.MustLoadFile(propPtr, properties.ISO_8859_1)

	// or through Decode
	var cfg ReportConfig
	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}
	cfg.File = fmt.Sprintf(cfg.File, "-"+"Inactive Users"+"-"+cfg.ProjectCategory)
	CreateInactiveUsersReport(cfg)
}

func addUser(name string, dispName string, projetcname string, active bool) {

	index := name + projetcname
	val, ok := allProjectUsers[index]

	if ok {
		val.user = name
		val.displayName = dispName
		val.projectName = projetcname
		val.active = val.active || active
		allProjectUsers[index] = val
	} else {
		var theProjectUSer ProjectUserType
		theProjectUSer.user = name
		theProjectUSer.displayName = dispName
		theProjectUSer.projectName = projetcname
		theProjectUSer.active = active
		allProjectUsers[index] = theProjectUSer
	}

}

func CreateInactiveUsersReport(cfg ReportConfig) {

	allProjectUsers = make(map[string]ProjectUserType)

	excelutils.NewFile()

	excelutils.SetCellFontHeader()
	excelutils.WiteCellln("Introduction")
	excelutils.WiteCellln("Please Do not edit this page!")
	excelutils.WiteCellln("This page is created by the User Report script: " + "https://github/perolo/jira-scripts" + "/" + "InactiveUserReport")
	t := time.Now()
	excelutils.WiteCellln("Created by: " + cfg.User + " : " + t.Format(time.RFC3339))
	excelutils.WiteCellln("")

	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(cfg.User),
		Password: strings.TrimSpace(cfg.Pass),
	}

	jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(cfg.Host))
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	excelutils.SetCellFontHeader2()
	excelutils.WiteCellln("Users and Permissions")
	excelutils.NextLine()
	excelutils.AutoFilterStart()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Project Name")
	excelutils.NextCol()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Permission Scheme")
	excelutils.NextCol()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Role")
	excelutils.NextCol()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Type")
	excelutils.NextCol()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Uname")
	excelutils.NextCol()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Display Name")
	excelutils.NextCol()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Admin")
	excelutils.SetCellStyleRotate()
	excelutils.NextCol()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Team")
	excelutils.SetCellStyleRotate()
	excelutils.NextCol()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Browse")
	excelutils.SetCellStyleRotate()
	excelutils.NextCol()

	excelutils.SetTableHeader()
	excelutils.WiteCell("User Active")
	excelutils.SetCellStyleRotate()
	excelutils.NextCol()

	excelutils.NextLine()

	//projects, _, err := jiraClient.Project.GetList()
	excelutils.Check(err)
	projects, _, err := jiraClient.Project.GetList()
	excelutils.Check(err)
	for _, project := range *projects {
		fmt.Printf("Project: %s \n", project.Key)

		if project.ProjectCategory.Name == cfg.ProjectCategory {
			fmt.Printf("Project name: %s Key: %s\n", project.Name, project.Key)
			projPerm, _, err2 := jiraClient.Project.GetPermissionScheme(project.Key)
			Utilities.Check(err2)

			if projPerm.Name == "Permission Scheme - Standard - Closed Down" || projPerm.Name == "Permission Scheme - Standard - Closing Down" || projPerm.Name == "Archived Projects - Permission Scheme" {
				fmt.Printf("   Skipping project due to Permission Scheme\n")
			} else {
				roles, _, err := jiraClient.Role.GetRolesForProjectWithContext(context.Background(), project.Key)
				Utilities.Check(err)
				for _, arole := range *roles {
					//projRole, _, err := jiraClient.User.GetProjectRole(arole)
					projRole, _, err := jiraClient.Role.GetActorsForProjectRoleWithContext(context.Background(), project.Key, arole.ID)
					Utilities.Check(err)
					fmt.Printf("   Role: %s\n", arole.Name)

					for _, actor := range projRole.Actors {

						fmt.Printf("    Actor: %s\n", actor.Name)
						if actor.Name == "c-johlan" {
							fmt.Printf("   What?: %v\n", actor.Name)
						}
						if actor.Type == "atlassian-group-role-actor" {

						} else if actor.Type == "atlassian-user-role-actor" {

							usr, _, _ := jiraClient.User.Get(actor.Name)
							if usr != nil {
								if !usr.Active {
									addUser(usr.Name, usr.DisplayName, project.Name, usr.Active)
								}
							} else {
								addUser(actor.Name, actor.DisplayName, project.Name, false)

							}
							//addUser(project, projRole, member.Name, member.DisplayName, actor.Name, allProjectUsers, member.EmailAddress)
						} else {
							// QUE???
							Utilities.Check(nil)
						}
					}
				}
			}
		}
	}
	for _, user := range allProjectUsers {
		fmt.Printf("User : %s \n", user.user)
		//		excelutils.WiteCellnc(k)
		excelutils.WiteCellnc(user.projectName)
		excelutils.WiteCellnc(user.permissionScheme)
		excelutils.WiteCellnc(user.role)
		excelutils.WiteCellnc(user.group)
		excelutils.WiteCellnc(user.user)
		excelutils.WiteCellnc(user.displayName)
		excelutils.WiteBoolCellnc(user.adminPermission)
		excelutils.WiteBoolCellnc(user.teamPermission)
		excelutils.WiteBoolCellnc(user.browsePermission)
		excelutils.WiteBoolCellnc(user.active)
		excelutils.NextLine()
	}

	excelutils.AutoFilterEnd()

	excelutils.SetColWidth("A", "A", 60)
	//	excelutils.SetColWidth("B", "D", 30)
	//	excelutils.SetColWidth("E", "R", 5)
	// Save xlsx file by the given path.
	excelutils.SaveAs(cfg.File)
	if cfg.Report {
		var config = client.ConfluenceConfig{}
		var copt client.OperationOptions
		config.Username = cfg.User
		config.Password = cfg.Pass
		config.URL = cfg.ConfHost
		config.Debug = false
		confluenceClient := client.Client(&config)

		copt.Title = "Inactive Users Report"
		copt.SpaceKey = "AAAD"

		_, name := filepath.Split(cfg.File)
		err = utilities.AddAttachmentAndUpload(confluenceClient, copt, name, cfg.File, "Created by Inactive Users Report")
		if err!= nil {
			panic(err)
		}
	}
}
