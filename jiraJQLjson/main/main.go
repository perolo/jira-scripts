package main

import (
	"flag"
	"github.com/perolo/jira-scripts/jiraJQLjson"
)

func main() {
	propPtr := flag.String("prop", "../../jiracategory.properties", "a string")
	flag.Parse()
	jirajqljson.JiraJQLjson(*propPtr)
}
