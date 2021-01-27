package main

import (
	"flag"
	"github.com/perolo/jira-scripts/projectpermissionsreport"
)

func main() {
	propPtr := flag.String("prop", "confluence.properties", "a string")

	projectpermissionsreport.ProjectPermissionsReport(*propPtr)
}
