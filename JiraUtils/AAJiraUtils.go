package JiraUtils

import (
	"fmt"
	"github.com/perolo/jira-client"
)

func AddComment(jiraClient *jira.Client, issue jira.Issue, comment string ) error {
	i := make(map[string]interface{})
	fields := make(map[string]interface{})
	afield := make(map[string]interface{})
	bfield := make(map[string]interface{})
	bfield["body"] = comment
	afield["add"] = bfield
	var vfield []map[string]interface{}

	fields["comment"] = append(vfield, afield)
	i["update"] = fields

	resp, err := jiraClient.Issue.UpdateIssue(issue.ID, i)

	if (err!=nil) {
		fmt.Printf("StatusCode: %v err: %s \n", resp.StatusCode, err.Error())
	} else {
		fmt.Printf("StatusCode: %v \n", resp.StatusCode)
	}
	return err

}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func AddLabelIfMissing(issue jira.Issue, label string, jiraClient *jira.Client) error {
	var  err error
	if Contains(issue.Fields.Labels, label) {
		// Skip
		fmt.Printf("Label already set: %s \n", issue.Key)
	} else {
		err = AddComment(jiraClient, issue, "User found Deactivated in AD")
		if err != nil {
			panic(err)
		}
		err = AddLabel(jiraClient, issue, label)
		if err != nil {
			panic(err)
		}
	}
	return err
}

func AddLabel(jiraClient *jira.Client, issue jira.Issue, label string ) error {
	i := make(map[string]interface{})
	fields := make(map[string]interface{})
	afield := make(map[string]interface{})
	//bfield := make(map[string]interface{})
	//bfield["items"] = label
	afield["add"] = label
	var vfield []map[string]interface{}

	fields["labels"] = append(vfield, afield)
	i["update"] = fields

	resp, err := jiraClient.Issue.UpdateIssue(issue.ID, i)

	if (err!=nil) {
		fmt.Printf("StatusCode: %v err: %s \n", resp.StatusCode, err.Error())
	} else {
		fmt.Printf("StatusCode: %v \n", resp.StatusCode)
	}
	return err

}


func SetSummary(jiraClient *jira.Client, issue jira.Issue, newSummary string) error {
	//Modify Summary
	i := make(map[string]interface{})
	fields := make(map[string]interface{})
	afield := make(map[string]interface{})
	afield["set"] = newSummary
	var vfield []map[string]interface{}

	fields["summary"] = append(vfield, afield)
	i["update"] = fields

	resp, err := jiraClient.Issue.UpdateIssue(issue.ID, i)

	if (err!=nil) {
		fmt.Printf("StatusCode: %v err: %s \n", resp.StatusCode, err.Error())
	} else {
		fmt.Printf("StatusCode: %v \n", resp.StatusCode)
	}
	return err
}

func RemoveComponent(jiraClient *jira.Client, issue jira.Issue, emeta *jira.EditMetaInfo) error{
	i := make(map[string]interface{})
	fields := make(map[string]interface{})
	afield := make(map[string]interface{})
	comp := emeta.Fields["components"]
	fmt.Printf("comp: %s\n", comp)
	v, ok := comp.(map[string]interface{})
	if !ok {
		// Can't assert, handle error.
		return  fmt.Errorf("Component: Illegal name")
	}
	d := v["allowedValues"].([]interface{})
	for _, s := range d {
		dd := s.(map[string]interface{})
		fmt.Printf("Value: %v\n", dd["name"])
		if dd["name"] == issue.Fields.Components[0].Name {
			afield["remove"] = dd
		}
	}
	var vfield []map[string]interface{}

	fields["components"] = append(vfield, afield)
	i["update"] = fields
	/*
		b, err := json.Marshal(i)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return;
		}
		fmt.Println(string(b))
	*/
	resp, err := jiraClient.Issue.UpdateIssue(issue.ID, i)

	if (err!=nil) {
		fmt.Printf("StatusCode: %v err: %s \n", resp.StatusCode, err.Error())
	} else {
		fmt.Printf("StatusCode: %v \n", resp.StatusCode)
	}
	return err
}

//"customfield_10515" = User
func SetUser(jiraClient *jira.Client, issue jira.Issue, newUser string) error{
	i := make(map[string]interface{})
	fields := make(map[string]interface{})
	afield := make(map[string]interface{})
	bfield := make(map[string]interface{})
	bfield["name"] = newUser
	afield["set"] = bfield
	var vfield []map[string]interface{}

	fields["customfield_10515"] = append(vfield, afield)
	i["update"] = fields

	resp, err := jiraClient.Issue.UpdateIssue(issue.ID, i)

	if (err!=nil) {
		fmt.Printf("StatusCode: %v err: %s \n", resp.StatusCode, err.Error())
	} else {
		fmt.Printf("StatusCode: %v \n", resp.StatusCode)
	}
	return err
}

//"customfield_10515" = User
func SetNewName(jiraClient *jira.Client, issue jira.Issue, newUser string) error{
	i := make(map[string]interface{})
	fields := make(map[string]interface{})
	afield := make(map[string]interface{})
	afield["set"] = newUser
	var vfield []map[string]interface{}

	fields["customfield_10712"] = append(vfield, afield)
	i["update"] = fields

	resp, err := jiraClient.Issue.UpdateIssue(issue.ID, i)

	if (err!=nil) {
		fmt.Printf("StatusCode: %v err: %s \n", resp.StatusCode, err.Error())
	} else {
		fmt.Printf("StatusCode: %v \n", resp.StatusCode)
	}
	return err
}
//"customfield_16410" = Email
func SetNewEmail(jiraClient *jira.Client, issue jira.Issue, email string) error{
	i := make(map[string]interface{})
	fields := make(map[string]interface{})
	afield := make(map[string]interface{})
	afield["set"] = email
	var vfield []map[string]interface{}

	fields["customfield_16410"] = append(vfield, afield)
	i["update"] = fields

	resp, err := jiraClient.Issue.UpdateIssue(issue.ID, i)

	if (err!=nil) {
		fmt.Printf("StatusCode: %v err: %s \n", resp.StatusCode, err.Error())
	} else {
		fmt.Printf("StatusCode: %v \n", resp.StatusCode)
	}
	return err
}


