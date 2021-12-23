package projectgroupconsolidator

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/excel-utils"
	"github.com/perolo/jira-client"
	"github.com/perolo/jira-scripts/jirautils"
	"log"
	"net/url"
	"os"
	"strings"
)

type ReportConfig struct {
	JiraHost   string `properties:"jirahost"`
	JiraUser   string `properties:"jirauser"`
	JiraPass   string `properties:"jirapass"`
	JiraToken  string `properties:"jiratoken"`
	UseToken   bool   `properties:"usetoken"`
	Simple     bool   `properties:"simple"`
	Archivedwf string `properties:"archivedwf"`
}

type ProjectGroupType struct {
	projectName string
	Group       string
	Role        string
}

/*
var permissions = []string{
	"PROJECT_ADMIN",
	"ASSIGNABLE_USER",
	"BROWSE",
}
*/

func ProjectGroupConsolidator(propPtr string) {

	flag.Parse()

	p := properties.MustLoadFile(propPtr, properties.ISO_8859_1)

	// or through Decode
	var cfg ReportConfig
	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}
	if cfg.UseToken {
		cfg.JiraPass = cfg.JiraToken
	} else {
	}

	if cfg.JiraHost != "https://jira-dev.assaabloy.net" {
		log.Fatal(nil)
	}

	if cfg.Simple {

	} else {

		tp := jira.BasicAuthTransport{
			Username: strings.TrimSpace(cfg.JiraUser),
			Password: strings.TrimSpace(cfg.JiraPass),
			UseToken: cfg.UseToken,
		}

		jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(cfg.JiraHost))
		if err != nil {
			fmt.Printf("\nerror: %v\n", err)
			return
		}
		if cfg.UseToken {
			jiraClient.Authentication.SetTokenAuth(cfg.JiraToken, cfg.UseToken)
		} else {
			jiraClient.Authentication.SetBasicAuth(cfg.JiraUser, cfg.JiraPass, cfg.UseToken)
		}
		roleLookup := make(map[string]string)

		for _, group := range Groups {
			fmt.Printf("Category: %s \n", group)
			project, _, _ := jiraClient.Project.Get(group.projectName)

			projPerm, closedDown := jirautils.GetPermissionScheme(jiraClient, project.Key, cfg.Archivedwf)

			if closedDown {
				fmt.Printf("   Skipping project due to Permission Scheme\n")
			} else {
				fmt.Printf("    Permission Scheme: %s\n", projPerm)
				// Validate that group and Role is correct in Permissionscheme

				// Get list of permissions for role

				// Get all members in Group

				// Get all users in Project
				//   if permissions as user
				//      if user not in group
				//          add to group
				//          querry OK to remove as user in project?
				//             if ok remove permission
				allUsers := make(map[string]string)
				//members, ret, err := jiraClient.Group.Get(group.Group)

				//safe := url.QueryEscape(group.Group)
				contjira := true
				startjira := 0
				maxjira := 50
				for contjira {

					jiramembers, _, _ := jiraClient.Group.GetWithOptionsWithContext(context.Background(), group.Group, &jira.GroupSearchOptions{StartAt: startjira, MaxResults: maxjira})

					for _, jiramember := range jiramembers {
						if _, ok := allUsers[jiramember.Name]; ok {
						} else {
							allUsers[jiramember.Name] = jiramember.Name
						}
					}
					if len(jiramembers) != maxjira {
						contjira = false
					} else {
						startjira = startjira + maxjira
					}
				}

				roles, _, err := jiraClient.Role.GetRolesForProjectWithContext(context.Background(), project.Key)

				excelutils.Check(err)
				for _, arole := range *roles {
					roleLookup[arole.ID] = arole.Name
					if arole.Name == group.Role {
						projRole, _, err := jiraClient.Role.GetActorsForProjectRoleWithContext(context.Background(), project.Key, arole.ID)
						excelutils.Check(err)
						fmt.Printf("   Role: %s\n", arole.Name)

						for _, actor := range projRole.Actors {

							if actor.Type == "atlassian-group-role-actor" {
								fmt.Printf("   %s atlassian-group-role-actor: %s\n", arole.Name, actor.Name)

							} else if actor.Type == "atlassian-user-role-actor" {

								if _, ok := allUsers[actor.Name]; ok {
									reader := bufio.NewReader(os.Stdin)

									fmt.Printf("Already Member   %s atlassian-group-role-actor: %s\n", arole.Name, actor.Name)

									fmt.Printf("Remove user: %s from role: %s in project: %s ?\n", actor.Name, arole.Name, group.projectName)

									fmt.Printf("Remove [y/n]: ")

									response, err := reader.ReadString('\n')
									if err != nil {
										log.Fatal(err)
									}

									response = strings.ToLower(strings.TrimSpace(response))

									if response == "y" || response == "yes" {

										fmt.Printf("Removing user: %s to role: %s in project: %s ?\n", actor.Name, arole.Name, group.projectName)
										_, _, err := jiraClient.Role.RemoveUserActorsForProjectRole(project.Key, projRole.ID, actor.Name)

										if err != nil {
											fmt.Printf("Failed to remove user. Group: %s status: %s \n", group.Group, actor.Name)

										}
									}

								} else {

									fmt.Printf("Move user: %s into group: %s in project: %s ?\n", actor.Name, group.Group, group.projectName)
									reader := bufio.NewReader(os.Stdin)

									fmt.Printf("Move [y/n]: ")

									response, err := reader.ReadString('\n')
									if err != nil {
										log.Fatal(err)
									}

									response = strings.ToLower(strings.TrimSpace(response))

									if response == "y" || response == "yes" {

										safe := url.QueryEscape(group.Group)
										fmt.Printf("Adding user: %s to role: %s in project: %s ?\n", actor.Name, arole.Name, group.projectName)
										_, _, err = jiraClient.Group.Add(safe, actor.Name)
										if err != nil {
											fmt.Printf("Failed to remove user. Group: %s status: %s \n", group.Group, actor.Name)

										}

										fmt.Printf("Removing user: %s to role: %s in project: %s ?\n", actor.Name, arole.Name, group.projectName)
										_, _, err = jiraClient.Role.RemoveUserActorsForProjectRole(project.Key, projRole.ID, actor.Name)

										if err != nil {
											fmt.Printf("Failed to add user. Group: %s status: %s \n", group.Group, actor.Name)
										}
									}

								}

							}
						}
					}
				}
				// Loop through all Roles with Permissions
				theprojPerm, _, _ := jiraClient.Project.GetProjectPermissions(project.Key)

				allRoles := make(map[string]string)

				for _, perm := range theprojPerm.Permissions {

					if perm.Holder.Type == "projectRole" {
						allRoles[roleLookup[perm.Holder.Parameter]] = roleLookup[perm.Holder.Parameter]

					}
				}

				for _, arole := range *roles {
					if _, ok := allRoles[arole.ID]; ok {
						fmt.Printf("Role: %s has a permission in scheme: %s \n", arole.Name, projPerm)
					} else {
						fmt.Printf("Role: %s has no permissions in scheme: %s \n", arole.Name, projPerm)
					}
				}
			}

		}
	}
}
