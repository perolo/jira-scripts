package main

import (
	"flag"
	"github.com/perolo/jira-scripts/syncjiraadgroup"
)

func main() {
	propPtr := flag.String("prop", "confluence.properties", "a properties file")

	syncjiraadgroup.JiraSyncAdGroup(*propPtr)
}
