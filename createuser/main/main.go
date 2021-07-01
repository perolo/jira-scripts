package main

import (
	"flag"
	"github.com/perolo/jira-scripts/createuser"
)

func main() {
	propPtr := flag.String("prop", "../../createuser.properties", "a string")
	flag.Parse()
	createuser.CreateUser(*propPtr)
}
