package syncjiraadgroup

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/confluence-client/client"
	"github.com/perolo/confluence-scripts/utilities"
	"github.com/perolo/excel-utils"
	"github.com/perolo/jira-client"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host            string `properties:"jirahost"`
	ConfHost        string `properties:"confhost"`
	JiraUser        string `properties:"jirauser"`
	ConfUser        string `properties:"confuser"`
	ConfPass        string `properties:"confpass"`
	JiraPass        string `properties:"jirapass"`
	JiraToken       string `properties:"jiratoken"`
	ConfToken       string `properties:"conftoken"`
	UseToken        bool   `properties:"usetoken"`
	Simple          bool   `properties:"simple"`
	AddOperation    bool   `properties:"add"`
	RemoveOperation bool   `properties:"remove"`
	AutoDisable     bool   `properties:"autodisable"`
	Report          bool   `properties:"report"`
	Limited         bool   `properties:"limited"`
	AdGroup         string `properties:"adgroup"`
	Localgroup      string `properties:"localgroup"`
	File            string `properties:"file"`
	Reset           bool   `properties:"reset"`
	//	Reset            bool `properties:"reset"`
	ConfUpload   bool   `properties:"confupload"`
	ConfPage     string `properties:"confluencepage"`
	ConfSpace    string `properties:"confluencespace"`
	ConfAttName  string `properties:"conlfuenceattachment"`
	Bindusername string `properties:"bindusername"`
	Bindpassword string `properties:"bindpassword"`
	BaseDN       string `properties:"basedn"`
}

func initReport(cfg Config) {
	if cfg.Report {
		excelutils.NewFile()
		excelutils.SetCellFontHeader()
		excelutils.WiteCellln("Introduction")
		excelutils.WiteCellln("Please Do not edit this page!")
		excelutils.WiteCellln("This page is created by the projectreport script: github.com\\perolo\\confluence-scripts\\SyncADGroup")
		t := time.Now()
		excelutils.WiteCellln("Created by: " + cfg.ConfUser + " : " + t.Format(time.RFC3339))
		excelutils.WiteCellln("")
		excelutils.WiteCellln("The Report Function shows:")
		excelutils.WiteCellln("   AdNames - Name and user found in AD Group")
		excelutils.WiteCellln("   JIRA Users - Name and user found in JIRA Group")
		excelutils.WiteCellln("   Not in AD - Users in the Local Group not found in the AD")
		excelutils.WiteCellln("   Not in JIRA - Users in the AD not found in the JIRA Group")
		excelutils.WiteCellln("   AD Errors - Internal error when searching for user in AD")
		excelutils.WiteCellln("")

		excelutils.SetCellFontHeader2()
		excelutils.WiteCellln("Group Mapping")
		excelutils.WriteColumnsHeaderln([]string{"AD Group", "Local group", "Add", "Remove", "Ad Count", "Local Count"})
		if cfg.Simple {
			excelutils.WriteColumnsln([]string{cfg.AdGroup, cfg.Localgroup, strconv.FormatBool(cfg.AddOperation), strconv.FormatBool(cfg.RemoveOperation)})
		} else {
			for _, syn := range GroupSyncs {
				if syn.InJira {
					excelutils.WriteColumnsln([]string{syn.AdGroup, syn.LocalGroup, excelutils.BoolToEmoji(syn.DoAdd), excelutils.BoolToEmoji(syn.DoRemove)})
				}
			}
		}
		excelutils.WiteCellln("")
		excelutils.SetCellFontHeader2()
		excelutils.WiteCellln("Report")

		excelutils.AutoFilterStart()
		var headers = []string{"Report Function", "AD group", "Local Group", "Name", "Uname", "Mail", "Error", "DN"}
		excelutils.WriteColumnsHeaderln(headers)
	}
}

func endReport(cfg Config) error {
	if cfg.Report {
		file := fmt.Sprintf(cfg.File, "-JIRA")
		excelutils.SetAutoColWidth()
		excelutils.SetColWidth("A", "A", 50)
		excelutils.AutoFilterEnd()
		excelutils.SaveAs(file)

		if cfg.ConfUpload {

			var config = client.ConfluenceConfig{}
			var copt client.OperationOptions
			config.Username = cfg.ConfUser
			config.Password = cfg.ConfPass
			config.UseToken = cfg.UseToken
			config.URL = cfg.ConfHost
			config.Debug = false
			confluenceClient := client.Client(&config)

			// Intentional override
			copt.Title = "Using AD groups for JIRA/Confluence"
			copt.SpaceKey = "AAAD"
			_, name := filepath.Split(file)
			cfg.ConfAttName = name
			return utilities.AddAttachmentAndUpload(confluenceClient, copt, name, file, "Created by Sync AD group")

		}

	}
	return nil
}

func JiraSyncAdGroup(propPtr string) {
	//	propPtr := flag.String("prop", "confluence.properties", "a string")
	flag.Parse()
	p := properties.MustLoadFile(propPtr, properties.ISO_8859_1)
	var cfg Config
	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}
	// Temporary workaround solution - need to find better?
	if cfg.UseToken {
		cfg.ConfPass = cfg.ConfToken
		cfg.JiraPass = cfg.JiraToken
	} else {
	}

	toolClient := toollogin(cfg)
	initReport(cfg)
	adutils.InitAD(cfg.Bindusername, cfg.Bindpassword)
	x := 15
	if cfg.Simple {
		SyncGroupInTool(cfg, toolClient)
	} else {
		for _, syn := range GroupSyncs {
			// If this is enabled the reports are partial
			//			if schedulerutil.CheckScheduleDetail(fmt.Sprintf("JiraSyncAdGroup-%s", syn.LocalGroup), time.Hour*24, cfg.Reset, schedulerutil.DummyFunc, "jiracategory.properties") {
			adCount := 0
			groupCount := 0
			if !syn.InJira && !syn.InConfluence {
				log.Fatal("Error in setup")
			}
			if syn.InJira {
				cfg.AdGroup = syn.AdGroup
				cfg.Localgroup = syn.LocalGroup
				cfg.AddOperation = syn.DoAdd
				cfg.RemoveOperation = syn.DoRemove
				cfg.AutoDisable = syn.AutoDisable
				adCount, groupCount = SyncGroupInTool(cfg, toolClient)
				excelutils.SetCell(fmt.Sprintf("%v", adCount), 5, x)
				excelutils.SetCell(fmt.Sprintf("%v", groupCount), 6, x)
				if adCount == groupCount {
					excelutils.SetCellBackground("#CCFFCC", 5, x)
					excelutils.SetCellBackground("#CCFFCC", 6, x)
				}
			} // Dirty Solution - find a better?
		}
		x = x + 1
	}
	//	}
	err := endReport(cfg)
	if err != nil {
		panic(err)
	}
	adutils.CloseAD()
}

func toollogin(cfg Config) *jira.Client {
	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(cfg.JiraUser),
		Password: strings.TrimSpace(cfg.JiraPass),
		UseToken: cfg.UseToken,
	}
	jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(cfg.Host))

	//jiraClient.Debug = true
	if cfg.UseToken {
		jiraClient.Authentication.SetTokenAuth(cfg.JiraToken, cfg.UseToken)
	} else {
		jiraClient.Authentication.SetBasicAuth(cfg.JiraUser, cfg.JiraPass, cfg.UseToken)
	}
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return nil
	}
	return jiraClient
}

func SyncGroupInTool(cfg Config, client *jira.Client) (adcount int, localcount int) {
	var toolGroupMemberNames map[string]adutils.ADUser
	fmt.Printf("\n")
	fmt.Printf("SyncGroup Jira AdGroup: %s LocalGroup: %s \n", cfg.AdGroup, cfg.Localgroup)
	fmt.Printf("\n")
	var adUnames []adutils.ADUser
	if cfg.AdGroup != "" {
		adUnames, _ = adutils.GetUnamesInGroup(cfg.AdGroup, cfg.BaseDN)
		fmt.Printf("adUnames(%v)\n", len(adUnames))
		if len(adUnames) == 0 {
			fmt.Printf("Warning empty AD group! adUnames(%v)\n", len(adUnames))
			panic(nil)
		}
	}
	if cfg.Report {
		if !cfg.Limited {
			for _, adu := range adUnames {
				var row = []string{"AD Names", cfg.AdGroup, cfg.Localgroup, adu.Name, adu.Uname, adu.Mail, adu.Err, adu.DN}
				excelutils.WriteColumnsln(row)
			}
		}
		adcount = len(adUnames)
	}
	if cfg.Localgroup != "" {
		toolGroupMemberNames = getUnamesInToolGroup(client, cfg.Localgroup)
		if cfg.Report {
			if !cfg.Limited {
				for _, tgm := range toolGroupMemberNames {
					var row = []string{"JIRA Users", cfg.AdGroup, cfg.Localgroup, tgm.Name, tgm.Uname, tgm.Mail, tgm.Err, tgm.DN}
					excelutils.WriteColumnsln(row)
				}
			}
		}
		localcount = len(toolGroupMemberNames)
	}
	if cfg.Localgroup != "" && cfg.AdGroup != "" {
		notInTool := adutils.Difference(adUnames, toolGroupMemberNames)
		if len(notInTool) == 0 {
			fmt.Printf("Not In Tool(%v)\n", len(notInTool))
		} else {
			fmt.Printf("Not In Tool(%v) ", len(notInTool))
			for _, nit := range notInTool {
				fmt.Printf("%s, ", nit.Uname)
			}
			fmt.Printf("\n")
		}

		if cfg.Report {
			for _, nji := range notInTool {
				var row = []string{"AD group users not found in Tool user group", cfg.AdGroup, cfg.Localgroup, nji.Name, nji.Uname, nji.Mail, nji.Err, nji.DN}
				excelutils.WriteColumnsln(row)
			}
		}
		notInAD := adutils.Difference2(toolGroupMemberNames, adUnames)
		if len(notInAD) == 0 {
			fmt.Printf("notInAD(%v)\n", len(notInAD))
		} else {
			fmt.Printf("notInAD(%v) ", len(notInAD))
			for _, nit := range notInAD {
				fmt.Printf("%s, ", nit.Uname)
			}
			fmt.Printf("\n")
		}
		if cfg.Report {
			for _, nad := range notInAD {
				if nad.DN == "" {
					dn, err := adutils.GetActiveUserDN(nad.Uname, cfg.BaseDN)
					if err == nil {
						nad.DN = dn.DN
						nad.Mail = dn.Mail
						nad.Name = dn.Name
					} else {
						udn, err := adutils.GetAllUserDN(nad.Uname, cfg.BaseDN)
						if err == nil {
							nad.DN = udn.DN
							nad.Mail = udn.Mail
							nad.Name = udn.Name
							nad.Err = "Deactivated"
							if cfg.AutoDisable == true {
								TryDeactivateUserJira(cfg.BaseDN, client, nad.Uname)
							}
						} else {
							edn, err := adutils.GetAllEmailDN(nad.Mail, cfg.BaseDN)
							if err == nil {
								nad.DN = edn[0].DN
								nad.Mail = edn[0].Mail
								nad.Err = edn[0].Err
								for _, ldn := range edn {
									var row2 = []string{"Tool user group member not found in AD group (multiple?)", cfg.AdGroup, cfg.Localgroup, nad.Name, nad.Uname, ldn.Mail, ldn.Err, ldn.DN}
									excelutils.WriteColumnsln(row2)
								}
							} else {
								nad.Err = err.Error()
							}
						}
					}
				}
				var row = []string{"Tool user group member not found in AD group", cfg.AdGroup, cfg.Localgroup, nad.Name, nad.Uname, nad.Mail, nad.Err, nad.DN}
				excelutils.WriteColumnsln(row)
			}
		}
		if cfg.AddOperation {
			for _, notin := range notInTool {
				if notin.Err == "" {
					fmt.Printf("Add user. Group: %s status: %s \n", cfg.Localgroup, notin.Uname)
					_, _, err := client.Group.Add(cfg.Localgroup, notin.Uname)
					if err != nil {
						fmt.Printf("Failed to add user. Group: %s status: %s \n", cfg.Localgroup, notin.Uname)
					}
				} else {
					fmt.Printf("Ad Problems skipping add: %s \n", notin.Uname)
				}

			}
		}
		if cfg.RemoveOperation {
			for _, notin := range notInAD {
				if notin.Err == "" {
					fmt.Printf("About to Remove user. Group: %s Uname: %s Name: %s \n", cfg.Localgroup, notin.Uname, notin.Name)
					fmt.Printf("Remove [y/n]: ")

					reader := bufio.NewReader(os.Stdin)
					response, err := reader.ReadString('\n')
					if err != nil {
						log.Fatal(err)
					}

					response = strings.ToLower(strings.TrimSpace(response))

					if response == "y" || response == "yes" {

						_, err := client.Group.Remove(cfg.Localgroup, notin.Uname)
						if err != nil {
							fmt.Printf("Failed to remove user. Group: %s status: %s \n", cfg.Localgroup, notin.Uname)
						}
					} else {
						fmt.Printf("Respone No - skipping remove: %s \n", notin.Uname)
					}
				}
			}
		}
	}
	return adcount, localcount
}

func getUnamesInToolGroup(theClient *jira.Client, localgroup string) map[string]adutils.ADUser {
	groupMemberNames := make(map[string]adutils.ADUser)
	cont := true
	start := 0
	max := 50
	for cont {
		groupMembers, resp, err := theClient.Group.GetWithOptions(localgroup, &jira.GroupSearchOptions{StartAt: start, MaxResults: max, IncludeInactiveUsers: false})
		if err != nil {
			if resp.StatusCode == 404 { // group not found?
				theClient.Group.AddGroup(localgroup)
				groupMembers, resp, err = theClient.Group.GetWithOptions(localgroup, &jira.GroupSearchOptions{StartAt: start, MaxResults: max, IncludeInactiveUsers: false})
				if err != nil {
					panic(err)
				}

			} else {
				panic(err)
			}
		}
		for _, member := range groupMembers {
			if _, ok := groupMemberNames[member.Name]; !ok {
				var newUser adutils.ADUser
				newUser.Uname = member.Name
				newUser.Name = member.DisplayName
				newUser.Mail = member.EmailAddress
				groupMemberNames[member.Name] = newUser
			}
		}
		if len(groupMembers) != max {
			cont = false
		} else {
			start = start + max
		}
	}
	return groupMemberNames
}

func DeactivateUser(jiraClient *jira.Client, user string) (*jira.UpdateResponse, *jira.Response, error) {
	i := make(map[string]interface{})
	i["active"] = false

	uresp, resp, err := jiraClient.User.Update(user, i)

	if err != nil {
		fmt.Printf("StatusCode: %v err: %s \n", resp.StatusCode, err.Error())
	} else {
		fmt.Printf("StatusCode: %v \n", resp.StatusCode)
	}
	return uresp, resp, err
}

var deactCounter = 0

func TryDeactivateUserJira(basedn string, client *jira.Client, deactuser string) {
	deactUser, _, err := client.User.Get(deactuser)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		if deactUser.Active == true {
			uresp, resp2, err := DeactivateUser(client, deactuser)
			deactCounter++
			if deactCounter > 10 {
				_, errn := adutils.GetAllUserDN("perolo", basedn)
				if errn != nil {
					fmt.Printf("Error: finding %s \n", "perolo")
					panic(errn)
				}
				deactCounter = 0
			}

			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			} else {
				if resp2.StatusCode == 400 {
					fmt.Printf("uresp.Name: %s\n", uresp.Name)
					fmt.Printf("uresp.Name: %v\n", resp2.StatusCode)
				}
				deactUser2, resp3, err := client.User.Get(deactuser)
				if resp3.StatusCode == 400 {
					fmt.Printf("uresp.Name: %s\n", uresp.Name)
					fmt.Printf("uresp.Name: %v\n", resp3.StatusCode)
				}
				if err == nil {
					if deactUser2.Active == false {
						fmt.Printf("deactUser: %s\n", deactUser2.Name)
						fmt.Printf("deactUser Active: %t\n", deactUser2.Active)
						fmt.Printf("respcode: %v\n", resp3.StatusCode)

					} else {
						fmt.Printf("Error: %s\n", err.Error())
						panic(err)
					}
				} else {
					fmt.Printf("Error: %s\n", err.Error())
					panic(err)
				}
			}
		}
	}
}
