package createuser

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/confluence-client/client"
	"github.com/perolo/jira-client"
	"github.com/perolo/myexcelize"
	"log"
	"net/http"
	//adutils "github.com/perolo/ad-utils"

	//	adutils "github.com/perolo/ad-utils"
	"strings"
)

type Config struct {
	JiraHost         string `properties:"jirahost"`
	ConfHost         string `properties:"confhost"`
	JiraUser         string `properties:"jirauser"`
	UseToken         bool   `properties:"usetoken"`
	JiraPass         string `properties:"jirapass"`
	JiraToken        string `properties:"jiratoken"`
	ConfUser         string `properties:"confuser"`
	ConfPass         string `properties:"confpass"`
	ConfToken        string `properties:"conftoken"`
	NewUser          string `properties:"new_user"`
	ConfluenceGroups string `properties:"new_confluence_groups"`
	JIRAGroups       string `properties:"new_jira_groups"`
	DisplayName      string `properties:"new_name"`
	PassWord         string `properties:"new_pass"`
	Email            string `properties:"new_email"`
	JIRA             bool   `properties:"jira"`
	Confluence       bool   `properties:"confluence"`
	//	Debug            bool   `properties:"debug"`
	Simple bool   `properties:"simple"`
	File   string `properties:"file"`
	//	CheckAD      bool   `properties:"checkad"`
	//	Bindusername string `properties:"bindusername"`
	//	Bindpassword string `properties:"bindpassword"`
	//	BaseDN       string `properties:"basedn"`
}

//var propConfig Config
var confluenceConfig = client.ConfluenceConfig{}

func CreateUser(propPtr string) {
	var err error
	var cfg Config
	var confClient *client.ConfluenceClient
	var jiraClient *jira.Client
	flag.Parse()

	p := properties.MustLoadFile(propPtr, properties.ISO_8859_1)

	if err = p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	// Start JIRA
	if cfg.JIRA {

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

		fmt.Printf("User: %s\n", cfg.NewUser)

	}

	// Start Confluence
	if cfg.Confluence {

		confluenceConfig.Username = cfg.ConfUser
		confluenceConfig.Password = cfg.ConfPass
		confluenceConfig.UseToken = cfg.UseToken
		confluenceConfig.URL = cfg.ConfHost
		//		confluenceConfig.Debug = propConfig.Debug

		confClient = client.Client(&confluenceConfig)

	}
	/* TODO Start AD
	if cfg.CheckAD {
		adutils.InitAD(cfg.Bindusername, cfg.Bindpassword)

	}

	*/
	if cfg.Simple {
		// TODO Manual Check? OK confirm?
		if doCreateUser(cfg, confClient, jiraClient) {
			return
		}
	} else {
		var err error
		fexcel, err := myexcelize.OpenFile(cfg.File)
		if err != nil {
			fmt.Printf("Result: %v\n", err.Error())
			return
		}

		rows, err := fexcel.GetRows("Sheet1")
		if err != nil {
			log.Fatal(err)
		}

		for _, row := range rows {
			cfg.NewUser = row[0]
			cfg.DisplayName = row[1]
			cfg.Email = row[2]
			cfg.PassWord = row[3]
			fmt.Printf("NewUser: %v\n", cfg.NewUser)
			if doCreateUser(cfg, confClient, jiraClient) {
				return
			}
		}
	}
	/*
		if cfg.CheckAD {
			adutils.CloseAD()
		}

	*/
}

func doCreateUser(propConfig Config, confClient *client.ConfluenceClient, jiraClient *jira.Client) bool {

	/*
		if propConfig.CheckAD {
			aduser, err := adutils.GetActiveUserDN(propConfig.NewUser, propConfig.BaseDN)
			if err != nil {
				//			fmt.Printf("\nerror: %v\n", err)
				fmt.Printf("User: %s not found in the Ad\n", propConfig.NewUser)
				reader := bufio.NewReader(os.Stdin)

				fmt.Printf("Create user anyway? [y/n]: ")

				response, err := reader.ReadString('\n')
				if err != nil {
					log.Fatal(err)
				}

				response = strings.ToLower(strings.TrimSpace(response))

				if response == "y" || response == "yes" {
					fmt.Printf("Attempting to create user even if not available in AD: %s\n", propConfig.NewUser)
					// do nothing continue
					//return false
				} else {
					fmt.Printf("Skipping to create user: %s\n", propConfig.NewUser)
					return false
				}
			} else {
				fmt.Printf("User: %s found in the Ad: %s\n", aduser.Name, aduser.DN)
			}
		}
	*/
	if propConfig.Confluence {
		err := createConfluenceUser(confClient, propConfig)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return true
		}
	}
	if propConfig.JIRA {
		err := createJiraUser(jiraClient, propConfig)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return true
		}
	}
	return false
}

func createConfluenceUser(confluence *client.ConfluenceClient, propConfig Config) error {
	var cUser *client.UserType
	var cResp *http.Response
	addGroups := propConfig.ConfluenceGroups != ""
	fmt.Printf("Checking if user exists in Confluence.\n")

	cUser, cResp = confluence.GetUser(propConfig.NewUser) //nolint:bodyclose
	defer jira.CleanupH(cResp)
	if cResp.StatusCode == 200 {
		fmt.Printf("User Already Exists in Confluence. %s\n", cUser.UserName)
	} else {
		fmt.Printf("Attempting to Create User in Confluence. %s\n", propConfig.NewUser)
		nUser := new(client.UserCreateType)

		nUser.UserName = propConfig.NewUser
		nUser.DisplayName = propConfig.DisplayName
		nUser.Password = propConfig.PassWord
		nUser.Email = propConfig.Email

		resp2 := confluence.CreateUser(*nUser) //nolint:bodyclose
		defer jira.CleanupH(resp2)
		if resp2.StatusCode == 200 {
			var dResp *http.Response
			fmt.Printf("User Created in Confluence. %s\n", nUser.UserName)
			cUser, dResp = confluence.GetUser(nUser.UserName) //nolint:bodyclose
			defer jira.CleanupH(dResp)
			if addGroups {
				addGroups = true
			}
		} else {
			fmt.Printf("Failed Creating USer in Confluence. %s\n", cUser.UserName)
			return fmt.Errorf("user not found %s", cUser.UserName)
		}

	}
	if addGroups {
		var users []string
		users = append(users, cUser.UserName)
		groups := strings.Split(propConfig.ConfluenceGroups, ",")
		for _, group := range groups {
			addUser := confluence.AddGroupMembers(group, users)
			if addUser.Status == "success" {
				fmt.Printf("Message: %s. User: %s added to %s \n", addUser.Message, users, group)
			} else {
				return fmt.Errorf("failed to add user %s to group %s", cUser.UserName, group)
			}
		}

	}
	return nil
}

func createJiraUser(jiraClient *jira.Client, propConfig Config) error {
	fmt.Printf("Checking if user exists in JIRA.\n")
	var usr *jira.User
	var resp *jira.Response
	var err error
	addGroups := propConfig.JIRAGroups != ""
	usr, resp, err = jiraClient.User.Get(propConfig.NewUser)

	if err != nil && resp != nil && resp.StatusCode == 404 {
		fmt.Printf("User not found in JIRA\n")
		fmt.Printf("Attempting to Create user in JIRA\n")
		user := new(jira.User)

		user.Name = propConfig.NewUser
		user.DisplayName = propConfig.DisplayName
		user.Password = propConfig.PassWord
		user.EmailAddress = propConfig.Email
		user.ApplicationKeys = []string{"jira-core"}

		//var buf io.ReadWriter
		buf := new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(user)
		if err != nil {
			fmt.Printf("\nerror: %v\n", err)
			return err
		}
		fmt.Printf("json: %s\n", buf)

		usr, resp, err = jiraClient.User.Create(user)
		if err != nil {
			fmt.Printf("\nerror: %v\n", err)
			return err
		}

		if resp.StatusCode == 201 {
			fmt.Printf("User created OK: %s\n", user.Name)
		} else {
			fmt.Printf("Failed to Create User: %s\n", user.Name)
			return err
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
				_, resg, err2 := jiraClient.Group.Add(group, usr.Name)
				if err2 != nil {
					fmt.Printf("error: %v\n", err2)
				} else {
					if resg.StatusCode == 201 {
						fmt.Printf("Added to group  %s \n", group)
					} else {
						fmt.Printf("Problem encoutered failed to add user to group. \n")
					}
				}
			} else {
				fmt.Printf("\nerror: %v\n", err)
				return err
			}
		}
	}
	return nil
}
