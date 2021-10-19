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
		JiraHost    string `properties:"host"`
		User        string `properties:"user"`
		Pass        string `properties:"password"`
		Source      string `properties:"source"`
		Destination string `properties:"destination"`
	}
	var cfg Config
	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}
	var err error
	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(cfg.User),
		Password: strings.TrimSpace(cfg.Pass),
	}

	jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(cfg.JiraHost))
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
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

