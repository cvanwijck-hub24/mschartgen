package mschartgen

import (
	//"fmt"
	"fmt"
	"io"
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
	err = os.WriteFile(rawJsonFile, orgJson, 0644)
	if err != nil {
		panic(err)
	}
}

func getRequest(url string, bearerToken string) (responseBody string, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", bearerToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error on response: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Debug("2xx response: ", resp.Status)
	} else {
		log.Errorf("non-2xx response code: %d - %s", resp.StatusCode, resp.Status)
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Response body: %s", string(body))
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func getOrgName(apiVersion string, bearerToken string) (orgName string, err error) {
	s, err := getRequest("https://graph.microsoft.com/"+apiVersion+"/organization", bearerToken)
	if err != nil {
		return "", err
	}
	data := Organization{}
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal organization data: %w", err)
	}

	return data.Data[0].DisplayName, nil
}

func getDirectReportsOfMember(memberId string) (members []Member, err error) {
	directReports, err := getRequest("https://graph.microsoft.com/"+apiVersion+"/users/"+memberId+"/directReports", bearerToken)
	if err != nil {
		return nil, err
	}
	data := DirectReports{}
	if err := json.Unmarshal([]byte(directReports), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal direct reports data: %w", err)
	}

	// iterate over each person, converting it into a member
	for _, v := range data.Data {
		members = append(members, Member{Id: v.Id, Name: v.Name, Title: v.Title, UserPrincipalName: v.UserPrincipalName})
	}

	return members, nil
}

func traverseTree(member Member) Member {
	members, err := getDirectReportsOfMember(member.Id)
	if err != nil {
		log.Errorf("Failed to get direct reports for member %s: %v", member.Id, err)
		return member
	}
	log.Debugf("Members: %+v", members)

	for i, directReport := range members {
		members[i] = traverseTree(directReport)
	}

	member.DirectReports = members
	return member
}
