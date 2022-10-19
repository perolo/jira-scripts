package main

import (
	"flag"

	"github.com/perolo/jira-scripts/projectreport"
)

func main() {
	propPtr := flag.String("prop", "../projectreport.properties", "a string")
	flag.Parse()

	projectreport.ProjectCategoryReport(*propPtr)
}
