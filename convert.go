package mschartgen

import (
	//	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func addChildren(parent *Child, directReportsMap map[string]Member) {
	directReports := make([]Member, 0, len(directReportsMap))
	for _, member := range directReportsMap {
		directReports = append(directReports, member)
	}

	for _, v := range directReports {
		if v.Name != parent.Name {
			child := Child{Name: v.Name, Title: v.Title}
			parent.Children = append(parent.Children, child)
			addChildren(&parent.Children[len(parent.Children)-1], directReportsMap)
		}
	}
}

func convert(org Organisation) {
	log.Debug("organisation pre-convert:", org)

	// start with the head of the org
	orgChart := OrgChart{}
	head := Child{Name: org.Head.Name, Title: org.Head.Title}
	orgChart.Data = append(orgChart.Data, head)

	// populate the org chart with the head's direct reports
	directReportsMap := make(map[string]Member)
	for _, member := range org.Head.DirectReports {
		directReportsMap[member.Name] = member
	}
	addChildren(&orgChart.Data[0], directReportsMap)

	// log the org chart before conversion to json for saving
	log.Debugf("OrgChart: %+v", orgChart)

	// render it to a json file for orgchartjs
	jsonData, err := json.MarshalIndent(orgChart.Data[0], "", "    ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("./srv/data.json", jsonData, 0666)
	if err != nil {
		panic(err)
	}
}
