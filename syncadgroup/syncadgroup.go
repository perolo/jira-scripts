package main

import (
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/ad-utils"
	excel_utils "github.com/perolo/excel-utils"
	"github.com/perolo/jira-client"
	"log"
	"strings"
	"time"
)

// or through Decode
type Config struct {
	Host         string `properties:"host"`
	User         string `properties:"user"`
	Pass         string `properties:"password"`
	AddOperation bool   `properties:"add"`
	Report 		 bool   `properties:"report"`
	ADgroup      string `properties:"adgroup"`
	Localgroup   string `properties:"localgroup"`
	File   		 string `properties:"file"`
	Bindusername string `properties:"bindusername"`
	Bindpassword string `properties:"bindpassword"`
}

func initReport(cfg Config) {
	if cfg.Report {
		excel_utils.NewFile()

		excel_utils.SetCellFontHeader()
		excel_utils.WiteCellln("Introduction")

		excel_utils.WiteCellln("Please Do not edit this page!")
		excel_utils.WiteCellln("This page is created by the projectreport script: github.com\\perolo\\confluence-scripts\\SyncADGroup")
		t := time.Now()

		excel_utils.WiteCellln("Created by: " + cfg.User + " : " + t.Format(time.RFC3339))
		excel_utils.WiteCellln("")
		excel_utils.WiteCellln("The Report Function shows:")
		excel_utils.WiteCellln("   AdNames - Name and user found in AD Group")
		excel_utils.WiteCellln("   JIRA Users - Name and user found in JIRA Group")
		excel_utils.WiteCellln("   Not in AD - Users in the Local Group not found in the AD")
		excel_utils.WiteCellln("   Not in JIRA - Users in the AD not found in the JIRA Group")
		excel_utils.WiteCellln("   AD Errors - Internal error when searching for user in AD")

		excel_utils.WiteCellln("")
		excel_utils.AutoFilterStart()
		var headers = []string{"Report Function", "AD group", "Local Group", "Name", "Uname", "Error"}
		excel_utils.WriteColumnsHeaderln(headers)

	}
}

func endReport(cfg Config) {
	if cfg.Report {

		file := fmt.Sprintf(cfg.File, "-JIRA")
		excel_utils.AutoFilterEnd()
		excel_utils.SaveAs(file)
	}
}
func main() {

	propPtr := flag.String("prop", "confluence.properties", "a string")

	flag.Parse()

	p := properties.MustLoadFile(*propPtr, properties.ISO_8859_1)

	var cfg Config
	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(cfg.User),
		Password: strings.TrimSpace(cfg.Pass),
	}

	jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(cfg.Host))
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}
	initReport(cfg)

	ad_utils.InitAD(cfg.Bindusername, cfg.Bindpassword)

	for _, syn := range GroupSyncs {
		//var adUnames []ad_utils.ADUser
		confGroupMemberNames := make(map[string]ad_utils.ADUser)
		cfg.ADgroup = syn.AdGroup
		cfg.Localgroup = syn.LocalGroup
		SyncGroupInConfluence(cfg, jiraClient, confGroupMemberNames)

	}
	endReport(cfg)
	ad_utils.CloseAD()
}


func main2() {

	propPtr := flag.String("prop", "confluence.properties", "a string")

	flag.Parse()

	p := properties.MustLoadFile(*propPtr, properties.ISO_8859_1)

	var cfg Config
	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}
	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(cfg.User),
		Password: strings.TrimSpace(cfg.Pass),
	}

	jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(cfg.Host))
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	ad_utils.InitAD(cfg.Bindusername, cfg.Bindpassword)
	initReport(cfg)

//	var adUnames []ad_utils.ADUser
	toolGroupMemberNames := make(map[string]ad_utils.ADUser)

	SyncGroupInConfluence(cfg, jiraClient, toolGroupMemberNames)
	endReport(cfg)
	ad_utils.CloseAD()
}

func SyncGroupInConfluence(cfg Config, client *jira.Client, toolGroupMemberNames map[string]ad_utils.ADUser) {
	fmt.Printf("\n")
	fmt.Printf("SyncGroup AdGroup: %s LocalGroup: %s \n", cfg.ADgroup, cfg.Localgroup)
	fmt.Printf("\n")
	var adUnames, aderrs []ad_utils.ADUser
	if cfg.ADgroup != ""  {
		adUnames, _, aderrs = ad_utils.GetUnamesInGroup(cfg.ADgroup)
		fmt.Printf("adUnames(%v): %s \n", len(adUnames), adUnames)
	}

	if cfg.Report {
		for _, adu := range adUnames {
//			var row = []string{"AD group", "group", "fun", "Name", "Uname"}
			var row = []string{"AD Names", cfg.ADgroup, cfg.Localgroup,  adu.Name, adu.Uname}
			excel_utils.WriteColumnsln(row)
		}
		for _, aderr := range aderrs {
			//			var row = []string{"AD group", "group", "fun", "Name", "Uname"}
			var row = []string{"AD Errors", cfg.ADgroup, cfg.Localgroup,  aderr.Name, aderr.Uname, aderr.Err}
			excel_utils.WriteColumnsln(row)
		}

	}
	if cfg.Localgroup != "" {
		getUnamesInJIRAGroup(client, cfg.Localgroup, toolGroupMemberNames)
		if (cfg.Report) {
			for _, tgm := range toolGroupMemberNames {
				//			var row = []string{"AD group", "group", "fun", "Name", "Uname"}
				var row = []string{"JIRA Users", cfg.ADgroup, cfg.Localgroup,  tgm.Name, tgm.Uname}
				excel_utils.WriteColumnsln(row)
			}
		}
	}

	if cfg.Localgroup != "" && cfg.ADgroup != "" {

		notInJIRA := ad_utils.Difference(adUnames, toolGroupMemberNames)
		fmt.Printf("notInJIRA(%v): %s \n", len(notInJIRA), notInJIRA)
		if cfg.Report {
			for _, nji := range notInJIRA {
				//			var row = []string{"AD group", "group", "fun", "Name", "Uname"}
				var row = []string{"Not in JIRA", cfg.ADgroup, cfg.Localgroup, nji.Name, nji.Uname}
				excel_utils.WriteColumnsln(row)
			}
		}

		notInAD := ad_utils.Difference2(toolGroupMemberNames, adUnames)
		fmt.Printf("notInAD: %s \n", notInAD)
		if cfg.Report {
			for _, nad := range notInAD {
				//			var row = []string{"AD group", "group", "fun", "Name", "Uname"}
				var row = []string{"Not in AD", cfg.ADgroup, cfg.Localgroup, nad.Name, nad.Uname}
				excel_utils.WriteColumnsln(row)
			}
		}

		if cfg.AddOperation {
			for _, user := range notInJIRA {

				fmt.Printf("Add user. Group: %s status: %s \n", cfg.Localgroup, user)
				_, _, err := client.Group.Add(cfg.Localgroup, user.Uname)
				if err != nil {
					fmt.Printf("Failed to add user. Group: %s status: %s \n", cfg.Localgroup, user)
				}
			}
		}
	}
}

func getUnamesInJIRAGroup(client *jira.Client, localgroup string, groupMemberNames map[string]ad_utils.ADUser) {

	cont := true
	start := 0
	max := 50
	for cont {

		jiraGroupMembers, _, err := client.Group.GetWithOptions(localgroup, &jira.GroupSearchOptions{StartAt: start, MaxResults: max})
		if err != nil {
			panic(err)
		}

		for _, jiramember := range jiraGroupMembers {
			if _, ok := groupMemberNames[jiramember.Name]; !ok {
				var newUser ad_utils.ADUser
				newUser.Name = jiramember.DisplayName
				newUser.Uname = jiramember.Name
				groupMemberNames[jiramember.Name] = newUser
			}
		}
		if len(jiraGroupMembers) != max {
			cont = false
		} else {
			start = start + max
		}
	}

}
