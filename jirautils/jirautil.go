package jirautils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

/*
func GetPermissionScheme(jiraClient *jira.Client, project jira.ProjectType) (string, bool) {
	// Permission Scheme
	projPerm, _, err2 := jiraClient.Project.GetPermissionScheme(project.Key)
	excelutils.Check(err2)
	closedDown := projPerm.Name == "Permission Scheme - Standard - Closed Down" || projPerm.Name == "Archived Projects - Permission Scheme"
	return projPerm.Name, closedDown
}
*/

func QueryUser(que string) bool {

	fmt.Print(que)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y" || response == "yes"
}
