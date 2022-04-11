package main

import (
	"flag"
	customfieldreport "github.com/perolo/jira-scripts/customfieldreport"
)

func main() {
	propPtr := flag.String("prop", "../customfieldreport.properties", "a string")
	flag.Parse()
	customfieldreport.CustomFieldReport(*propPtr)

}
