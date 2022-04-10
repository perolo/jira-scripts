package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/excel-utils"
	"github.com/perolo/jira-client"
	"log"
	"strings"
)

func main() {

	propPtr := flag.String("prop", "../confluence-scripts/confluence.properties", "a string")

	flag.Parse()

	p := properties.MustLoadFile(*propPtr, properties.ISO_8859_1)

	// or through Decode
	type Config struct {
		JiraHost    string `properties:"jirahost"`
		JiraUser    string `properties:"jirauser"`
		UseToken    bool   `properties:"usetoken"`
		JiraPass    string `properties:"jirapass"`
		JiraToken   string `properties:"jiratoken"`
		Source      string `properties:"source"`
		Destination string `properties:"destination"`
	}
	var cfg Config
	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}
	var err error
	var jiraClient *jira.Client
	if cfg.UseToken {
		tp := jira.BearerAuthTransport{
			Token: strings.TrimSpace(cfg.JiraToken),
		}
		jiraClient, err = jira.NewClient(tp.Client(), strings.TrimSpace(cfg.JiraHost))
	} else {
		tp := jira.BasicAuthTransport{
			Username: strings.TrimSpace(cfg.JiraUser),
			Password: strings.TrimSpace(cfg.JiraPass),
		}
		jiraClient, err = jira.NewClient(tp.Client(), strings.TrimSpace(cfg.JiraHost))
	}

	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}
	if cfg.UseToken {
		//jiraClient.Authentication.SetTokenAuth(cfg.JiraToken, cfg.UseToken)
		//jiraClient.Authentication.
	} else {
		jiraClient.Authentication.SetBasicAuth(cfg.JiraUser, cfg.JiraPass)
	}

	projects, _, err := jiraClient.Project.GetList()
	excelutils.Check(err)
	for _, project := range *projects {

		if strings.Contains(project.ProjectCategory.Name, "GTT") {

			fmt.Printf("Project: %s : %s\n", project.Key, project.Name)

			roles, _, err := jiraClient.Role.GetRolesForProjectWithContext(context.Background(), project.Key)
			excelutils.Check(err)
			for _, arole := range *roles {
				projRole, _, err := jiraClient.Role.GetActorsForProjectRoleWithContext(context.Background(), project.Key, arole.ID)
				excelutils.Check(err)
				//			fmt.Printf("   Role: %s\n", arole.Name)

				for _, actor := range projRole.Actors {

					if actor.Type == "atlassian-group-role-actor" {
						fmt.Printf("   %s atlassian-group-role-actor: %s\n", arole.Name, actor.Name)

					} else if actor.Type == "atlassian-user-role-actor" {

						//					fmt.Printf("   atlassian-user-role-actor: %v\n", actor.Name)

					} else {
						// QUE???
						excelutils.Check(nil)
					}
				}
			}
		}
	}
}
