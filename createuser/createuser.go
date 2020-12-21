package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/confluence-prop/client"
	"github.com/perolo/jira-client"
	"io"
	"log"
	"net/http"
	"strings"
)

// or through Decode
type Config struct {
	JiraHost         string `properties:"jirahost"`
	User             string `properties:"user"`
	Pass             string `properties:"password"`
	ConfHost         string `properties:"confhost"`
	NewUser          string `properties:"new_user"`
	ConfluenceGroups string `properties:"new_confluence_groups"`
	JIRAGroups       string `properties:"new_jira_groups"`
	DisplayName      string `properties:"new_name"`
	PassWord         string `properties:"new_pass"`
	Email            string `properties:"new_email"`
	JIRA             bool   `properties:"jira"`
	Confluence       bool   `properties:"confluence"`
	Debug            bool   `properties:"debug"`
}

var propConfig Config
var confluenceConfig = client.ConfluenceConfig{}

func main() {
	var err error
	addGroups := false

	propPtr := flag.String("prop", "createuser-gtt.properties", "a string")

	flag.Parse()

	p := properties.MustLoadFile(*propPtr, properties.ISO_8859_1)

	if err = p.Decode(&propConfig); err != nil {
		log.Fatal(err)
	}

	// Start JIRA
	if propConfig.JIRA {
		/*
			var jiraConfig = jira_client2.JIRAConfig{}
			jiraConfig.Username = propConfig.User
			jiraConfig.Password = propConfig.Pass
			jiraConfig.URL = propConfig.JiraHost
			jiraConfig.Debug = propConfig.Debug

			jiraClient = jira_client2.Client(&jiraConfig)
		*/
		tp := jira.BasicAuthTransport{
			Username: strings.TrimSpace(propConfig.User),
			Password: strings.TrimSpace(propConfig.Pass),
		}

		jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(propConfig.JiraHost))
		if err != nil {
			fmt.Printf("\nerror: %v\n", err)
			return
		}

		fmt.Printf("User: %s\n", propConfig.NewUser)

		fmt.Printf("Checking if user exists in JIRA.\n")
		var usr *jira.User
		var resp *jira.Response
		usr, resp, err = jiraClient.User.Get(propConfig.NewUser)

		if resp != nil && resp.StatusCode == 404 {
			fmt.Printf("Attempting to Create user in JIRA\n")
			//			user := new(jira_client2.CreateUser)
			user := new(jira.User)

			user.Name = propConfig.NewUser
			user.DisplayName = propConfig.DisplayName
			user.Password = propConfig.PassWord
			user.EmailAddress = propConfig.Email
			user.ApplicationKeys = []string{"jira-core"}

			var buf io.ReadWriter
			buf = new(bytes.Buffer)
			err = json.NewEncoder(buf).Encode(user)
			fmt.Printf("json: %s\n", buf)

			usr, resp, err = jiraClient.User.Create(user)
			if err != nil {
				fmt.Printf("\nerror: %v\n", err)
				return
			}

			if resp.StatusCode == 201 {
				fmt.Printf("User created OK: %s\n", user.Name)
				addGroups = true
			} else {
				fmt.Printf("Failed to Create User: %s\n", user.Name)
			}

		} else {
			if usr != nil {
				fmt.Printf("User exists: %s\n", usr.DisplayName)
			}

			addGroups = true
			fmt.Printf("Adding User to group: %s\n", propConfig.JIRAGroups)
		}

		if addGroups {
			groups := strings.Split(propConfig.JIRAGroups, ",")
			for _, group := range groups {
				if usr != nil {
					_, resg, err := jiraClient.Group.Add(group, usr.Name)
					if err != nil {
						fmt.Printf("\nerror: %v\n", err)
						return
					}
					if resg.StatusCode == 201 {
						fmt.Printf("Added to group  %s \n", group)
					} else {
						fmt.Printf("Problem.\n")
					}
				} else {
					fmt.Printf("\nerror: %v\n", err)
					return
				}
			}
		}
	}

	// Start Confluence
	if propConfig.Confluence {

		fmt.Printf("Checking if user exists in Confluence.\n")

		confluenceConfig.Username = propConfig.User
		confluenceConfig.Password = propConfig.Pass
		confluenceConfig.URL = propConfig.ConfHost
		confluenceConfig.Debug = propConfig.Debug

		confluence := client.Client(&confluenceConfig)
		var cUser *client.UserType
		var cResp *http.Response
		cUser, cResp = confluence.GetUser(propConfig.NewUser)

		if cResp.StatusCode == 200 {
			fmt.Printf("User Already Exists in Confluence. %s\n", cUser.UserName)
			addGroups = true
		} else {
			fmt.Printf("Attempting to Create User in Confluence. %s\n", propConfig.NewUser)
			nUser := new(client.UserCreateType)

			nUser.UserName = propConfig.NewUser
			nUser.DisplayName = propConfig.DisplayName
			nUser.Password = propConfig.PassWord
			nUser.Email = propConfig.Email

			resp2 := confluence.CreateUser(*nUser)
			if resp2.StatusCode == 200 {
				fmt.Printf("User Created in Confluence. %s\n", nUser.UserName)
				cUser, cResp = confluence.GetUser(nUser.UserName)
				addGroups = true
			} else {
				fmt.Printf("Failed Creating USer in Confluence. %s\n", cUser.UserName)
			}

		}
		if addGroups {
			var users []string
			users = append(users, cUser.UserName)
			groups := strings.Split(propConfig.ConfluenceGroups, ",")
			for _, group := range groups {
				addUser := confluence.AddGroupMembers(group, users)
				fmt.Printf("Message: %s. User: %s added to %s \n", addUser.Message, users,group)
			}

		}
	}
}
