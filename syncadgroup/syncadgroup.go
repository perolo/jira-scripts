package main

import (
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/perolo/ad-utils"
	"github.com/perolo/jira-client"
	"log"
	"strings"
)

type GroupSyncType struct {
	AdGroup    string
	LocalGroup string
}
var GroupSyncs = []GroupSyncType{
	{AdGroup: "AD Group 1", LocalGroup: "Local 1"},
	{AdGroup: "AD Group 2", LocalGroup: "Local 2"},
}

func difference(a []string, b map[string]string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
func difference2(a map[string]string, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// or through Decode
type Config struct {
	Host         string `properties:"host"`
	User         string `properties:"user"`
	Pass         string `properties:"password"`
	AddOperation bool   `properties:"add"`
	ADgroup      string `properties:"adgroup"`
	Localgroup   string `properties:"localgroup"`
	Bindusername string `properties:"bindusername"`
	Bindpassword string `properties:"bindpassword"`
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

	ad_utils.InitAD(cfg.Bindusername, cfg.Bindpassword)

	for _, syn := range GroupSyncs {
		var adUnames []string
		confGroupMemberNames := make(map[string]string)
		cfg.ADgroup = syn.AdGroup
		cfg.Localgroup = syn.LocalGroup
		SyncGroupInConfluence(adUnames, cfg, jiraClient, confGroupMemberNames)

	}

	ad_utils.CloseAD()
}

func main2() {

	propPtr := flag.String("prop", "../confluence.properties", "a string")

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

	var adUnames []string
	toolGroupMemberNames := make(map[string]string)

	SyncGroupInConfluence(adUnames, cfg, jiraClient, toolGroupMemberNames)
	ad_utils.CloseAD()
}

func SyncGroupInConfluence(adUnames []string, cfg Config, client *jira.Client, toolGroupMemberNames map[string]string) {
	fmt.Printf("\n")
	fmt.Printf("SyncGroup AdGroup: %s LocalGroup: %s \n", cfg.ADgroup, cfg.Localgroup)
	fmt.Printf("\n")
	adUnames, _ = ad_utils.GetUnamesInGroup(cfg.ADgroup)
	fmt.Printf("adUnames: %s \n", adUnames)

	getUnamesInJIRAGroup(client, cfg, toolGroupMemberNames)

	notInJIRA := difference(adUnames, toolGroupMemberNames)
	fmt.Printf("notInJIRA: %s \n", notInJIRA)

	notInAD := difference2(toolGroupMemberNames, adUnames)
	fmt.Printf("notInAD: %s \n", notInAD)

	if cfg.AddOperation {
		for _, user := range notInJIRA {

			fmt.Printf("Add user. Group: %s status: %s \n", cfg.Localgroup, user)
			_, _, err := client.Group.Add(cfg.Localgroup, user)
			if err != nil {
				fmt.Printf("Failed to add user. Group: %s status: %s \n", cfg.Localgroup, user)
			}
		}
	}
}

func getUnamesInJIRAGroup(client *jira.Client, cfg Config, groupMemberNames map[string]string) {

	cont := true
	start := 0
	max := 50
	for cont {

		jiraGroupMembers, _, err := client.Group.GetWithOptions(cfg.Localgroup, &jira.GroupSearchOptions{StartAt: start, MaxResults: max})
		if err != nil {
			panic(err)
		}

		for _, jiramember := range jiraGroupMembers {
			if _, ok := groupMemberNames[jiramember.Name]; !ok {
				groupMemberNames[jiramember.Name] = jiramember.Name
			}
		}
		if len(jiraGroupMembers) != max {
			cont = false
		} else {
			start = start + max
		}
	}

}
