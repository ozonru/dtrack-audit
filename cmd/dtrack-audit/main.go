package main

import (
	"flag"
	"fmt"
	"github.com/ozonru/dtrack-audit/internal/dtrack"
	"log"
	"os"
	"time"
)

func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {
	var inputFileName, projectId, apiKey, apiUrl string
	var syncMode bool
	var uploadResult dtrack.UploadResult
	var findings []dtrack.Finding
	var timeout int

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Send SBOM file to Dependency Track.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of program:\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&inputFileName, "i", "bom.xml", "Target SBOM file")
	flag.StringVar(&projectId, "p", os.Getenv("DTRACK_PROJECT_ID"), "Project ID (Required)")
	flag.StringVar(&apiKey, "k", os.Getenv("DTRACK_API_KEY"), "API Key (Required)")
	flag.StringVar(&apiUrl, "u", os.Getenv("DTRACK_API_URL"), "API URL (Required)")
	flag.BoolVar(&syncMode, "s", false, "Sync mode enabled (Upload SBOM file and waith for scan result)")
	flag.IntVar(&timeout, "t", 25, "Max timeout in second for polling API for project findings")
	flag.Parse()

	if projectId == "" || apiKey == "" || apiUrl == "" {
		flag.Usage()
		os.Exit(1)
	}

	apiClient := dtrack.ApiClient{ApiKey: apiKey, ApiUrl: apiUrl}
	uploadResult, err := apiClient.Upload(inputFileName, projectId)
	checkError(err)

	if uploadResult.Token != "" {
		log.Printf("SBOM file is successfully uploaded to DTrack API. Result token is %s\n", uploadResult.Token)
	}

	if uploadResult.Token != "" && syncMode {
		err := apiClient.PollTokenBeingProcessed(uploadResult.Token, time.After(time.Duration(timeout)*time.Second))
		checkError(err)
		findings, err = apiClient.GetFindings(projectId)
		checkError(err)
		fmt.Println(findings)
		if len(findings) > 0 {
			log.Fatal(fmt.Errorf("Vulnerabilities found!"))
		}

		findings, err := apiClient.GetFindings(projectId)
		checkError(err)
		if len(findings) > 0 {
			fmt.Printf("%d vulnerabilities found!\n\n", len(findings))
			for _, f := range findings {
				fmt.Printf(" > %s: %s\n", f.Vuln.Severity, f.Vuln.Title)
				fmt.Printf("   Component: %s %s\n", f.Comp.Name, f.Comp.Version)
				fmt.Printf("   More info: %s\n\n", apiClient.GetVulnViewUrl(f.Vuln))
			}
			os.Exit(1)
		}
	}
}
