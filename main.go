package main

import (
	"flag"
	"github.com/perolo/jira-scripts/cloneissues"
)

func main() {
	propPtr := flag.String("prop", "cloneissues/cloneissues.properties", "a string")
	flag.Parse()
	cloneissues.CloneIssues(*propPtr)
}
