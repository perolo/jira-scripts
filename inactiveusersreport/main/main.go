package main

import (
	"flag"
	"github.com/perolo/jira-scripts/inactiveusersreport"
)

func main() {
	propPtr := flag.String("prop", "jira.properties", "a string")

	inactiveusersreport.InactiveUserReport(*propPtr)
}
