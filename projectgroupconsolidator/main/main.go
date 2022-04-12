package main

import (
	"flag"
	"github.com/perolo/jira-scripts/projectgroupconsolidator"
)

func main() {
	propPtr := flag.String("prop", "../projectgroupconsolidator.properties", "a string")

	projectgroupconsolidator.ProjectGroupConsolidator(*propPtr)
}
