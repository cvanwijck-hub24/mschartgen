package mschartgen

import (
	//"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	//"github.com/davecgh/go-spew/spew"
)

const (
	depth = 1
)

var (
	count = 0
)

func renderRawOrgChart(org Organisation) {
	// render the raw structure to debug
	log.Debug("======raw-org-chart======")
	log.Debug(org)
	log.Debug("============")

	// process to raw json at this stage
	orgJson, err := json.MarshalIndent(org, "", "    ")
	if err != nil {
		panic(err)
	}
	log.Debug("======org-chart-json======")
	log.Debug(string(orgJson))
	log.Debug("============")
	// save the raw/native output to a local file
	err = ioutil.WriteFile(rawJsonFile, orgJson, 0644)
	if err != nil {
		panic(err)
	}
}

func getRequest(url string, bearerToken string) (responseBody string, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", bearerToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error on response.\n[ERRO] -", err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Debug("2xx response: ", resp.Status)
	} else {
		log.Errorf("non-2xx response code:", resp)
		os.Exit(1)
	}

	//log.Debugf("resp:", resp)
	body, _ := ioutil.ReadAll(resp.Body)
	//log.Debugf("body:", string(body))

	byt := []byte(body)
	//log.Debugf("byt:", string(byt))

	return string(byt), err
}

func getOrgName(apiVersion string, bearerToken string) (orgName string, err error) {
	s, err := getRequest("https://graph.microsoft.com/"+apiVersion+"/organization", bearerToken)
	if err != nil {
		panic(err)
	}
	data := Organization{}
	json.Unmarshal([]byte(s), &data)

	return data.Data[0].DisplayName, err
}

func getDirectReportsOfMember(memberId string) (members []Member, err error) {
	directReports, err := getRequest("https://graph.microsoft.com/"+apiVersion+"/users/"+memberId+"/directReports", bearerToken)
	if err != nil {
		panic(err)
	}
	data := DirectReports{}
	json.Unmarshal([]byte(directReports), &data)

	// iterate over each person, converting it into a member
	for _, v := range data.Data {
		members = append(members, Member{Id: v.Id, Name: v.Name, Title: v.Title})
	}

	return members, err
}

func traverseTree(member Member) Member {
	members, _ := getDirectReportsOfMember(member.Id)
	log.Debug("Members:", members)

	for i, directReport := range members {
		members[i] = traverseTree(directReport)
	}

	member.DirectReports = members
	return member
}
