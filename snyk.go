package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/michael-go/go-jsn/jsn"
)

var ORGS = map[string]string{
	"Billing":              "c8dfa46f-86c7-4e1f-ad03-8b54eaf7686a",
	"Classic":              "b7b7afcd-242a-4756-8c83-56e69211d148",
	"Data Engineering":     "47f1fc68-410b-49c9-8910-53ab433b22a3",
	"Emerging Products":    "878e064a-6ef9-4e0f-bde8-fff0d437470f",
	"ExistingProducts":     "1156cfed-f570-49f4-bbc5-db2b69387632",
	"Nova":                 "d871051b-e7fc-4c7b-b590-73482b3c3fb1",
	"Platform":             "b7ac3f09-0ca7-435a-96f2-012c0cc57a45",
	"Product Integrations": "d85b4772-a027-40d6-9d9a-6454461f3921",
	"Wordpress":            "3cdadf0a-1b0a-4074-8b75-c47f940e0b01",
}

func getOrgIds(flags flags, customDebug debug) (map[string]string, error) {
	verb := "GET"
	api_version := "2024-08-22"
	baseURL := flags.mandatoryFlags.endpointAPI + "/rest"
	orgsAPI := "/orgs?version=" + api_version + "&limit=100"
	orgs, err := makeSnykAPIRequest_REST(verb, baseURL, orgsAPI, flags.mandatoryFlags.apiToken, nil, customDebug)
	if err != nil {
		log.Printf("*** ERROR *** Could not list Orgs for endpoint %s\n", orgsAPI)
	}

	// if org is passed as a flag, then use it
	if len(flags.optionalFlags.orgID) != 0 {
		ORGS = map[string]string{
			"UNKNOWN": flags.optionalFlags.orgID,
		}
	}

	orgIDs := map[string]string{}
	for _, v := range ORGS {
		for _, org := range orgs {
			if org.K("id").String().Value == v {
				name := org.K("attributes").K("name").String().Value
				log.Printf("*** INFO *** Found Org %s with ID %s\n", name, v)
				orgIDs[name] = v
			}
		}
	}
	if len(orgIDs) == 0 {
		log.Printf("*** ERROR *** Could not find any Orgs for endpoint %s\n", orgsAPI)
		err = errors.New("Could not find any Orgs for endpoint " + orgsAPI)
	}
	return orgIDs, err
}

func getOrgProjects(orgId string, flags flags, customDebug debug) ([]jsn.Json, error) {
	verb := "GET"
	api_version := "2022-07-08~beta"

	baseURL := flags.mandatoryFlags.endpointAPI + "/rest"

	projectsAPI := "/orgs/" + orgId + "/projects?version=" + api_version + "&status=active&limit=100"
	if len(flags.optionalFlags.projectCriticality) > 0 || len(flags.optionalFlags.projectEnvironment) > 0 || len(flags.optionalFlags.projectLifecycle) > 0 {

		if len(flags.optionalFlags.projectCriticality) > 0 {
			projectsAPI += "&businessCriticality=" + strings.Replace(flags.optionalFlags.projectCriticality, ",", "%2C", -1)
		}

		if len(flags.optionalFlags.projectEnvironment) > 0 {
			projectsAPI += "&environment=" + strings.Replace(flags.optionalFlags.projectEnvironment, ",", "%2C", -1)
		}

		if len(flags.optionalFlags.projectLifecycle) > 0 {
			projectsAPI += "&lifecycle=" + strings.Replace(flags.optionalFlags.projectLifecycle, ",", "%2C", -1)
		}
	}

	var err error

	projectList, err := makeSnykAPIRequest_REST(verb, baseURL, projectsAPI, flags.mandatoryFlags.apiToken, nil, customDebug)
	if err != nil {
		filters := "projectCriticality: " + flags.optionalFlags.projectCriticality + "\n projectEnvironment: " + flags.optionalFlags.projectEnvironment + "\n projectLifecycle: " + flags.optionalFlags.projectLifecycle
		log.Printf("*** ERROR *** Could not list the Project(s) for endpoint %s\n Applied Filters: %s\n", projectsAPI, filters)
		errorMessage := fmt.Sprintf("Failure, Could not list the Project(s) for endpoint %s .\n Applied filters: %s\n", projectsAPI, filters)
		writeErrorFile("getOrgProjects", errorMessage, customDebug)
		err = errors.New(errorMessage)
	}

	return projectList, err
}

func getProjectsIds(orgId string, options flags, customDebug debug, notCreatedLogFile string) ([]string, error) {

	var projectIds []string
	if len(options.optionalFlags.projectID) == 0 {
		filters := "projectCriticality: " + options.optionalFlags.projectCriticality + "\n projectEnvironment: " + options.optionalFlags.projectEnvironment + "\n projectLifecycle: " + options.optionalFlags.projectLifecycle
		log.Println("*** INFO *** Project ID not specified - listing all projects that match the following filters: ", filters)

		projects, err := getOrgProjects(orgId, options, customDebug)
		if err != nil {
			message := fmt.Sprintf("error while getting projects ID for org %s", options.optionalFlags.orgID)
			writeErrorFile("getProjectsIds", message, customDebug)
			return nil, err
		}

		for _, project := range projects {
			projectID := project.K("id").String().Value
			projectIds = append(projectIds, projectID)
		}

		if len(projectIds) == 0 {
			ErrorMessage := fmt.Sprintf("Failure, Could not retrieve project ID")
			writeErrorFile("getProjectsIds", ErrorMessage, customDebug)
			return projectIds, errors.New(ErrorMessage)
		}
		return projectIds, nil
	}

	projectIds = append(projectIds, options.optionalFlags.projectID)

	return projectIds, nil
}

func getProjectDetails(orgID string, Mf MandatoryFlags, projectID string, customDebug debug) (jsn.Json, error) {
	responseData, err := makeSnykAPIRequest("GET", Mf.endpointAPI+"/v1/org/"+orgID+"/project/"+projectID, Mf.apiToken, nil, customDebug)
	if err != nil {
		log.Printf("*** ERROR *** Could not get the Project detail for endpoint %s\n", Mf.endpointAPI)
		errorMessage := fmt.Sprintf("Failure, Could not get the Project detail for endpoint %s\n", Mf.endpointAPI)
		err = errors.New(errorMessage)
		writeErrorFile("getProjectDetails", errorMessage, customDebug)
	}

	project, err := jsn.NewJson(responseData)
	if err != nil {
		errorMessage := fmt.Sprintf("Failure, Could not read the Project detail for endpoint %s\n", Mf.endpointAPI)
		err = errors.New(errorMessage)
		writeErrorFile("getProjectDetails", errorMessage, customDebug)
	}

	return project, err
}
