# jira-scripts
Handy utilities for administration and maintenance of Jira
The use of properties files is to enable automation with CI tool without "logging" passwords.

## Clone Group Permissions
Confluence does not support renaming of groups.
This script clones the permissions of a group in all Spaces to a new group.


#### How to use
* Create the group(copygroupname) in Jira
* Modify the properties file
* Run the script
  ```
  go run clonegrouppermissions.go -prop clonegrouppermissions.properties
  ```
* (Optional) Remove the old group

The parameters needed are defined in the properties file:
```
confhost=http://jira.com:8080
user=user
password=password
source=origingroupname
destination=copygroupname
```
## Clone Issues
Clone all issues found by jql. Create a issue link between them and add a label.

#### Example of use
All Requirements and tasks needed to implement for a standard may be defined in a jira project and 
when implemented in a product cloned into that product (example GDPR)
The standard may be revsion controlled with the fixversion in Jira, and what products have implemented them in using 
the link.
And the progress of implementing a standard id managed in the project as any task.


#### How to use
* Create the jql in Jira
* Modify the properties file
* Run the script
  ```
  go run cloneissues.go -prop cloneissues.properties
  ```

The parameters needed are defined in the properties file:
```
jirahost=http://jira.com:8080
user=user
password=password
destinationproject=
jql=
clonesubtasks=
label=
```

## Create User
Create a user in Jira and/or Confluence add default groups to the user.
<p>
With Simple=false and users from an Excel sheet many users may be created at once
</p>
<p>
With checkad=true a check that the user exists in Active Directory may performed before adding user. 
</p>

#### Example of use
Enables automation of user creation and adding of permissions without user logging in.

#### How to use
* Modify the properties file
* Run the script
  ```
  go run createuser.go -prop createuser.properties
  ```

The parameters needed are defined in the createuser file:
```
jirahost=http://jira.com:8080
user=adminuser
password=admipassword
confhosthttp://confluence.com:8080
new_user=dummyuser
new_name=Dummy User
new_pass=dummypassword
new_email=dummy.user@nowhere.com
new_confluence_groups=confluence-users,project-1
new_jira_groups=jira-users,project-1
jira=true/false
confluence=true/false
simple=true/false
file=
checkad=true/false
bindusername=aduser
bindpassword=adpassword
basedn=dc=nowhere,dc=global
```

## Custom Field Report
An analysis report of the custom fields in Jira


## Description of analysis
The Report generates 4 different "Report Types" - some fields are common, other are not. 
Some fields may be used to get an overview of that field or project.

Please double check the result before acting on data.

Custom fields are either "Global" - or are restricted to a list of projects.

Projects using Permission schemes : "Permission Scheme - Standard - Closed Down" || "Archived Projects - Permission Scheme" - Are considered as Archived.

## Report sequence
1. Loop Through all Custom fields
   1. If Global Custom Field, ADD LINE IN EXCEL **"Report Type"** = **"Custom Field Global"**
      1. Search for Issues with that Field not EMPTY
         1. Loop through all those Projects
            1. ADD LINE IN EXCEL **"Report Type"** = **"Custom Field Use"**
   2. if Not Global Custom Field, ADD LINE IN EXCEL **"Report Type"** = **"Custom Field Project"**
      1. Loop Through all Projects
ADD LINE IN EXCEL "Report Type" = "Project"
Search if that Project has any Dedicated Custom Fields declared (Currently Maximum 5 projects added! - Performance optimization)
ADD LINE IN EXCEL "Report Type" = "Project Use"
Analysis / Things to investigate


#### Example of use
Enables automation of user creation and adding of permissions without user logging in.

#### How to use
* Modify the properties file
* Run the script
  ```
  go run createuser.go -prop createuser.properties
  ```

The parameters needed are defined in the createuser file:
```
jirahost=http://jira.com:8080
user=adminuser
password=admipassword
confhosthttp://confluence.com:8080
new_user=dummyuser
new_name=Dummy User
new_pass=dummypassword
new_email=dummy.user@nowhere.com
new_confluence_groups=confluence-users,project-1
new_jira_groups=jira-users,project-1
jira=true/false
confluence=true/false
simple=true/false
file=
checkad=true/false
bindusername=aduser
bindpassword=adpassword
basedn=dc=nowhere,dc=global
```

