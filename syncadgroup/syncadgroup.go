package syncadgroup

import (
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/ad-utils"
	"github.com/perolo/confluence-prop/client"
	"github.com/perolo/confluence-scripts/utilities"
	excelutils "github.com/perolo/excel-utils"
	"github.com/perolo/jira-client"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// or through Decode
type Config struct {
	Host            string `properties:"host"`
	ConfHost        string `properties:"confhost"`
	User            string `properties:"user"`
	Pass            string `properties:"password"`
	Simple          bool   `properties:"simple"`
	AddOperation    bool   `properties:"add"`
	RemoveOperation bool   `properties:"remove"`
	Report          bool   `properties:"report"`
	Limited         bool   `properties:"limited"`
	AdGroup         string `properties:"adgroup"`
	Localgroup      string `properties:"localgroup"`
	File            string `properties:"file"`
	ConfUpload      bool   `properties:"confupload"`
	ConfPage        string `properties:"confluencepage"`
	ConfSpace       string `properties:"confluencespace"`
	ConfAttName     string `properties:"conlfuenceattachment"`
	Bindusername    string `properties:"bindusername"`
	Bindpassword    string `properties:"bindpassword"`
}

func initReport(cfg Config) {
	if cfg.Report {
		excelutils.NewFile()
		excelutils.SetCellFontHeader()
		excelutils.WiteCellln("Introduction")
		excelutils.WiteCellln("Please Do not edit this page!")
		excelutils.WiteCellln("This page is created by the projectreport script: github.com\\perolo\\confluence-scripts\\SyncADGroup")
		t := time.Now()
		excelutils.WiteCellln("Created by: " + cfg.User + " : " + t.Format(time.RFC3339))
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
		if cfg.Simple {
			excelutils.WriteColumnsHeaderln([]string{"AD Group", "Local group", "Add", "Remove"})
			excelutils.WriteColumnsln([]string{cfg.AdGroup, cfg.Localgroup, strconv.FormatBool(cfg.AddOperation), strconv.FormatBool(cfg.RemoveOperation)})
		} else {
			excelutils.WriteColumnsHeaderln([]string{"AD Group", "Local group", "Add", "Remove"})
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
		excelutils.SetColWidth("A", "A", 60)
		excelutils.AutoFilterEnd()
		excelutils.SaveAs(file)

		if cfg.ConfUpload {

			var config = client.ConfluenceConfig{}
			var copt client.OperationOptions
			config.Username = cfg.User
			config.Password = cfg.Pass
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
	toolClient := toollogin(cfg)
	initReport(cfg)
	adutils.InitAD(cfg.Bindusername, cfg.Bindpassword)
	if cfg.Simple {
		SyncGroupInTool(cfg, toolClient)
	} else {
		for _, syn := range GroupSyncs {
			if !syn.InJira && !syn.InConfluence {
				log.Fatal("Error in setup")
			}
			if syn.InJira {
				cfg.AdGroup = syn.AdGroup
				cfg.Localgroup = syn.LocalGroup
				cfg.AddOperation = syn.DoAdd
				cfg.RemoveOperation = syn.DoRemove
				SyncGroupInTool(cfg, toolClient)
			}
		}
	}
	endReport(cfg)
	adutils.CloseAD()
}

func toollogin(cfg Config) *jira.Client {
	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(cfg.User),
		Password: strings.TrimSpace(cfg.Pass),
	}
	jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(cfg.Host))
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return nil
	}
	return jiraClient
}

func SyncGroupInTool(cfg Config, client *jira.Client) {
	var toolGroupMemberNames map[string]adutils.ADUser
	fmt.Printf("\n")
	fmt.Printf("SyncGroup AdGroup: %s LocalGroup: %s \n", cfg.AdGroup, cfg.Localgroup)
	fmt.Printf("\n")
	var adUnames []adutils.ADUser
	if cfg.AdGroup != "" {
		adUnames, _ = adutils.GetUnamesInGroup(cfg.AdGroup)
		fmt.Printf("adUnames(%v)\n", len(adUnames))
	}
	if cfg.Report {
		if !cfg.Limited {
			for _, adu := range adUnames {
				var row = []string{"AD Names", cfg.AdGroup, cfg.Localgroup, adu.Name, adu.Uname, adu.Mail, adu.Err, adu.DN}
				excelutils.WriteColumnsln(row)
			}
		}
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
					dn, err := adutils.GetActiveUserDN(nad.Uname)
					if err == nil {
						nad.DN = dn.DN
						nad.Mail = dn.Mail
					} else {
						udn, err := adutils.GetAllUserDN(nad.Uname)
						if err == nil {
							nad.DN = udn.DN
							nad.Mail = udn.Mail
							nad.Err = "Deactivated"
						} else {
							edn, err := adutils.GetAllEmailDN(nad.Mail)
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
					fmt.Printf("Remove user. Group: %s status: %s \n", cfg.Localgroup, notin.Uname)
					_, err := client.Group.Remove(cfg.Localgroup, notin.Uname)
					if err != nil {
						fmt.Printf("Failed to remove user. Group: %s status: %s \n", cfg.Localgroup, notin.Uname)
					}
				} else {
					fmt.Printf("Ad Problems skipping remove: %s \n", notin.Uname)
				}
			}
		}
	}
}

func getUnamesInToolGroup(theClient *jira.Client, localgroup string) map[string]adutils.ADUser {
	groupMemberNames := make(map[string]adutils.ADUser)
	cont := true
	start := 0
	max := 50
	for cont {
		groupMembers, _, err := theClient.Group.GetWithOptions(localgroup, &jira.GroupSearchOptions{StartAt: start, MaxResults: max})
		if err != nil {
			panic(err)
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
