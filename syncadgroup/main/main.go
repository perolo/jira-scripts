package main

import (
	"flag"
	"github.com/perolo/jira-scripts/syncadgroup"
)

func main() {
	propPtr := flag.String("prop", "gitlabmergestatus.properties", "a properties file")

	syncadgroup.JiraSyncAdGroup(*propPtr)
}
