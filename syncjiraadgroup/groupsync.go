package syncjiraadgroup

//import "github.com/perolo/jira-utils/shared"

//import "git.aa.st/perolo/jira-utils/shared"

//type GroupSyncType shared.GroupSyncType
//var GroupSyncs = shared.GroupSyncs

type GroupSyncType struct {
	AdGroup      string
	LocalGroup   string
	DoAdd        bool
	DoRemove     bool
	InJira       bool
	InConfluence bool
	AutoDisable  bool
}

var GroupSyncs = []GroupSyncType{
	{AdGroup: "AD Group 1", LocalGroup: "Local 1"},
	{AdGroup: "AD Group 2", LocalGroup: "Local 2"},
}
