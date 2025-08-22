package mschartgen

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	log "github.com/sirupsen/logrus"

	// "io/ioutil"
	"os"
)

const (
	//apiVersion = "v1.0"
	apiVersion  = "beta"
	jsonFile    = "./srv/data.json"
	rawJsonFile = "./srv/data.raw.json"
)

var (
	json               = jsoniter.ConfigCompatibleWithStandardLibrary
	bearerToken        = "Bearer " + os.Getenv("MSGRAPH_BEARER_TOKEN")
	org                = Organisation{}
	httpdListenPort    = 8080
	httpdRootDirectory = "./srv"
)

func Process(organisationName string, rootMemberId string) {
	// early validation
	if len(os.Getenv("MSGRAPH_BEARER_TOKEN")) < 1 {
		log.Fatal("error: first, set your bearer access token in the environment variable, MSGRAPH_BEARER_TOKEN")
	}

	// set the org name from upstream or if the user has specified an override
	if len(organisationName) < 1 {
		// get org name from msgraph
		orgName, err := getOrgName(apiVersion, bearerToken)
		if err != nil {
			panic(err)
		}
		log.Debugf("orgName: %s", orgName)
		// set the organisation name
		org.Name = orgName
	} else {
		org.Name = organisationName
	}

	// start with the head, usually the CEO
	rootPerson, err := getRequest("https://graph.microsoft.com/"+apiVersion+"/users/"+rootMemberId, bearerToken)
	if err != nil {
		panic(err)
	}
	log.Debugf("rootPerson: %s", rootPerson)
	person := Person{}
	if err := json.Unmarshal([]byte(rootPerson), &person); err != nil {
		panic(err)
	}
	log.Debug("----")
	log.Debugf("person: %+v", person)
	log.Debugf("Id: %s", person.Id)
	org.Head.Id = person.Id
	log.Debugf("displayName: %s", person.Name)
	org.Head.Name = person.Name
	log.Debugf("jobTitle: %s", person.Title)
	org.Head.Title = person.Title
	log.Debugf("UPN: %s", person.UserPrincipalName)
	org.Head.UserPrincipalName = person.UserPrincipalName
	log.Debugf("DirectReports: %+v", person.DirectReports)
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
	output, _ := json.Marshal(person)
	// decode it back to get a map
	var a interface{}
	json.Unmarshal(output, &a)
	b := a.(map[string]interface{})
	// Replace the map key
	b["name"] = b["displayName"]
	b["title"] = b["jobTitle"]
	delete(b, "displayName")
	delete(b, "jobTitle")

	// get direct reports
	// directReports, err := getDirectReportsOfMember(rootMemberId)
	// org.Head.DirectReports = directReports

	// direct reports traversal
	// do an each on the direct reports, appending to
	// for k, report := range org.Head.DirectReports {
	// 	r, err := getDirectReportsOfMember(report.Id)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	org.Head.DirectReports[k].DirectReports = r
	// 	//members = append(members, Member{Id: v.Id, Name: v.Name, Title: v.Title})
	// }

	//	traverseDirectReports(org.Head.Id, org)

	// now work on the org chart in preparation for render and convert

	// temp: this should contain the entire tree of direct reports
	org.Head = traverseTree(org.Head)

	// renders the raw org chart to file and log
	renderRawOrgChart(org)

	// process native into format used by upstream js lib
	// the method should output to srv/data.json itself
	//convert(org)

	// start the http server
	//Serve(httpdRootDirectory, httpdListenPort)
}
