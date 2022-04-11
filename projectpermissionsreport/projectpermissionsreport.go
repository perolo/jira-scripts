package projectpermissionsreport

import (
	"context"
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/excel-utils"
	"github.com/perolo/jira-client"
	"github.com/perolo/jira-scripts/jirautils"
	"log"
	"strings"
	"time"
)

type ReportConfig struct {
	JiraHost         string `properties:"jirahost"`
	JiraUser         string `properties:"jirauser"`
	JiraPass         string `properties:"jirapass"`
	JiraToken        string `properties:"jiratoken"`
	ConfHost         string `properties:"confhost"`
	ConfUser         string `properties:"confuser"`
	ConfPass         string `properties:"confpass"`
	ConfToken        string `properties:"conftoken"`
	UseToken         bool   `properties:"usetoken"`
	ProjectCategory  string `properties:"projectcategory"`
	ConfUpload       bool   `properties:"confupload"`
	File             string `properties:"file"`
	Simple           bool   `properties:"simple"`
	Report           bool   `properties:"report"`
	RolesReport      bool   `properties:"rolesreport"`
	ExpandGroups     bool   `properties:"expandgroups"`
	PermissionReport bool   `properties:"permissionreport"`
	Archivedwf       string `properties:"archivedwf"`
}

type ProjectUserType struct {
	projectName      string
	projectLead      string
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

var permissions = []string{
	"PROJECT_ADMIN",
	"ASSIGNABLE_USER",
	"BROWSE",
}

var allProjectUsers map[string]ProjectUserType

func ProjectPermissionsReport(propPtr string) {

	flag.Parse()

	p := properties.MustLoadFile(propPtr, properties.ISO_8859_1)

	// or through Decode
	var cfg ReportConfig
	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}
	if cfg.Simple {
		cfg.File = fmt.Sprintf(cfg.File, "-"+cfg.ProjectCategory)
		CreateProjectPermissionsReport(cfg)
	} else {
		reportBase := cfg.File
		/* All - too large....
		cfg.ProjectCategory = ""
		cfg.File = fmt.Sprintf(reportBase, "-all")
		fmt.Printf("Category: all \n")
		CreateProjectPermissionsReport(cfg)
		*/
		for _, category := range Categories {
			cfg.ProjectCategory = category
			cfg.File = fmt.Sprintf(reportBase, "-"+category)
			fmt.Printf("Category: %s \n", category)
			CreateProjectPermissionsReport(cfg)
		}
	}
}

func addUser(project jira.ProjectType, projLead string, projRole string, name string, dispName string, group string, scheme string, browse bool, team bool, admin bool, active bool) {

	var index = project.Name + "-" + name + "-" + group
	val, ok := allProjectUsers[index]

	if ok {
		val.user = name
		val.projectName = project.Name
		val.projectLead = projLead
		val.group = group
		if !strings.Contains(val.role, projRole) {
			val.role = val.role + " + " + projRole
		}
		val.permissionScheme = scheme
		val.displayName = dispName
		val.browsePermission = val.browsePermission || browse
		val.teamPermission = val.teamPermission || team
		val.adminPermission = val.adminPermission || admin
		val.active = val.active || active
		allProjectUsers[index] = val
	} else {
		var theProjectUSer ProjectUserType
		theProjectUSer.user = name
		theProjectUSer.displayName = dispName
		theProjectUSer.projectName = project.Name
		theProjectUSer.group = group
		if !strings.Contains(theProjectUSer.role, projRole) {
			theProjectUSer.role = theProjectUSer.role + " + " + projRole
		}
		theProjectUSer.permissionScheme = scheme
		theProjectUSer.browsePermission = browse
		theProjectUSer.teamPermission = team
		theProjectUSer.adminPermission = admin
		theProjectUSer.active = active
		allProjectUsers[index] = theProjectUSer
	}

}

func CreateProjectPermissionsReport(cfg ReportConfig) { //nolint:funlen
	var jiraClient *jira.Client
	var err error

	allProjectUsers = make(map[string]ProjectUserType)

	excelutils.NewFile()

	excelutils.SetCellFontHeader()
	excelutils.WiteCellln("Introduction")
	excelutils.WiteCellln("Please Do not edit this page!")
	excelutils.WiteCellln("This page is created by the User Report script: " + "https://github/perolo/jira-scripts" + "/" + "ProjectPermissionsReport")
	t := time.Now()
	excelutils.WiteCellln("Created by: " + cfg.ConfUser + " : " + t.Format(time.RFC3339))
	excelutils.WiteCellln("")

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

	excelutils.SetCellFontHeader2()
	excelutils.WiteCellln("Users and Permissions")
	excelutils.NextLine()
	excelutils.AutoFilterStart()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Project Name")
	excelutils.NextCol()

	excelutils.SetTableHeader()
	excelutils.WiteCell("Project Lead")
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

	projects, _, err := jiraClient.Project.GetList()
	excelutils.Check(err)
	for _, project := range *projects {
		if (cfg.ProjectCategory == "") || (project.ProjectCategory.Name == cfg.ProjectCategory) {

			projPerm, closedDown := jirautils.GetPermissionScheme(jiraClient, project.Key, cfg.Archivedwf)

			if closedDown {
				fmt.Printf("   Skipping project due to Permission Scheme\n") // mainly performance improvement, we know only admin can view
			} else {

				p, _, _ := jiraClient.Project.Get(project.ID)
				if cfg.RolesReport {
					roles, _, err := jiraClient.Role.GetRolesForProjectWithContext(context.Background(), project.Key)
					excelutils.Check(err)
					for _, arole := range *roles {
						//projRole, _, err := jiraClient.User.GetProjectRole(arole)
						projRole, _, err := jiraClient.Role.GetActorsForProjectRoleWithContext(context.Background(), project.Key, arole.ID)
						excelutils.Check(err)
						fmt.Printf("   Role: %s\n", arole.Name)

						for _, actor := range projRole.Actors {

							fmt.Printf("    Actor: %s\n", actor.Name)
							if actor.Name == "c-johlan" {
								fmt.Printf("   What?: %v\n", actor.Name)
							}
							if actor.Type == "atlassian-group-role-actor" {

								if cfg.ExpandGroups {

									cont := true
									start := 0
									max := 50
									for cont {

										//members, _, _, _ := jiraClient.Group.GetUsersFromGroup(safe, &jira.GroupOptions{StartAt: start, MaxResults: max})
										members, _, err := jiraClient.Group.GetWithOptionsWithContext(context.Background(), actor.Name, &jira.GroupSearchOptions{StartAt: start, MaxResults: max})
										excelutils.Check(err)
										for _, member := range members {
											addUser(project, p.Lead.Name, projRole.Name, member.Name, member.DisplayName, actor.Name, projPerm, false, false, false, false)
										}
										if len(members) != max {
											cont = false
										} else {
											start = start + max
										}
									}
								} else {
									addUser(project, p.Lead.Name, projRole.Name, actor.Name, actor.DisplayName, "group", projPerm, false, false, false, false)
								}
							} else if actor.Type == "atlassian-user-role-actor" {

								addUser(project, p.Lead.Name, projRole.Name, actor.Name, actor.DisplayName, "user", projPerm, false, false, false, false)
								//addUser(project, projRole, member.Name, member.DisplayName, actor.Name, allProjectUsers, member.EmailAddress)
							} else {
								// QUE???
								excelutils.Check(nil)
							}
						}
					}

				}
				if cfg.PermissionReport {
					//Loop through permissions
					for _, perm := range permissions {
						cont := true
						start := 0
						max := 50
						for cont {
							members, _, err := jiraClient.Group.SearchPermissionsWithOptionsWithContext(context.Background(), &jira.PermissionSearchOptions{StartAt: start, MaxResults: max, ProjectKey: project.Key, Permissions: perm})
							excelutils.Check(err)
							if members != nil {
								for _, mem := range *members {
									fmt.Printf("Permissions: %s User: %s\n", perm, mem.Name)

									addUser(project, p.Lead.Name, "PermSearch", mem.Name, mem.DisplayName, "user", projPerm, perm == permissions[2], perm == permissions[1], perm == permissions[0], mem.Active)
								}
								if len(*members) != max {
									cont = false
								} else {
									start = start + max
								}
							}
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
		excelutils.WiteCellnc(user.projectLead)
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

	excelutils.SetAutoColWidth()
	excelutils.AutoFilterEnd()
	excelutils.SetColWidth("A", "A", 60)

	//	excelutils.SetColWidth("A", "A", 40)
	//	excelutils.SetColWidth("B", "D", 30)
	//	excelutils.SetColWidth("E", "R", 5)
	// Save xlsx file by the given path.
	excelutils.SaveAs(cfg.File)
	/*
		if cfg.ConfUpload {
			if cfg.Report {
				var config = client.ConfluenceConfig{}
				config.Username = cfg.ConfUser
				config.Password = cfg.ConfPass
				config.UseToken = cfg.UseToken
				config.URL = cfg.ConfHost
				config.Debug = false
				confluenceClient := client.Client(&config)
				var copt client.OperationOptions
				copt.Title = "Project Permissions Reports"
				copt.SpaceKey = "AAAD"
				_, name := filepath.Split(cfg.File)
				err = utilities.AddAttachmentAndUpload(confluenceClient, copt, name, cfg.File, "Created by Project Permissions Report")
				if err != nil {
					panic(err)
				}
			}
		}

	*/
}
