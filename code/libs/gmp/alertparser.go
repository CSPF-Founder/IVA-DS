package gmp

import (
	"bytes"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
)

type Alert struct {
	IP                   string  `csv:"IP"`
	Hostname             string  `csv:"Hostname"`
	Severity             string  `csv:"Severity"`
	NVTName              string  `csv:"NVT Name"`
	NVTOID               string  `csv:"NVT OID"`
	Port                 string  `csv:"Port"`
	PortProto            string  `csv:"Port Protocol"`
	CVSS                 float64 `csv:"CVSS"`
	Summary              string  `csv:"Summary"`
	VulnerabilityInsight string  `csv:"Vulnerability Insight"`
	Impact               string  `csv:"Impact"`
	Solution             string  `csv:"Solution"`
	OtherReferences      string  `csv:"Other References"`
	AffectedSoftwareOrOS string  `csv:"Affected Software/OS"`
	CVEDetails           string  `csv:"CVEs"`
	SpecificResult       string  `csv:"Specific Result"`
}

// parseCSV parses the CSV data to a slice of Alert structs
func parseCSVData(csvData []byte, uniqueID string) ([]Alert, error) {
	tmpDir := os.TempDir()
	csvPath := tmpDir + "/report_" + uniqueID + ".csv"
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	file, err := os.Create(csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	pData := replaceStr(csvData)
	csvReader := csv.NewReader(bytes.NewReader(pData))
	csvReader.Comma = ','
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields
	csvReader.LazyQuotes = true

	alerts, err := parseToAlerts(csvReader)
	if err != nil {
		return nil, err
	}

	return alerts, nil
}

// parseToAlerts parses the CSV file to a slice of Alert structs
// This function reads the CSV file and maps the values to the Alert struct
//
//nolint:cyclop
func parseToAlerts(csvReader *csv.Reader) ([]Alert, error) {

	// Read headers - first row
	headers, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	var alerts []Alert

	// Create a map of headers -
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[header] = i
	}

	if len(headerMap) == 0 {
		return alerts, nil
	}

	// Iterate over each row and map values to the Alert struct
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) == 0 {
			continue
		}

		alert := Alert{}
		for header, index := range headerMap {
			if index >= len(record) {
				// Boundary check to avoid index out of range
				continue
			}

			value := strings.TrimSpace(record[index])
			switch header {
			case "IP":
				alert.IP = value
			case "Hostname":
				alert.Hostname = value
			case "Severity":
				alert.Severity = value
			case "NVT Name":
				alert.NVTName = value
			case "NVT OID":
				alert.NVTOID = value
			case "Port":
				alert.Port = value
			case "Port Protocol":
				alert.PortProto = value
			case "CVSS":
				alert.CVSS, _ = strconv.ParseFloat(value, 64)
			case "Summary":
				alert.Summary = value
			case "Vulnerability Insight":
				alert.VulnerabilityInsight = value
			case "Impact":
				alert.Impact = value
			case "Solution":
				alert.Solution = value
			case "Other References":
				alert.OtherReferences = value
			case "Affected Software/OS":
				alert.AffectedSoftwareOrOS = value
			case "CVEs":
				alert.CVEDetails = value
			case "Specific Result":
				alert.SpecificResult = value
			}
		}
		alerts = append(alerts, alert)
	}

	if len(alerts) == 0 {
		return alerts, nil
	}

	return filterAlerts(alerts), nil
}

func filterAlerts(
	alerts []Alert,
) []Alert {

	blacklistedOIDList := map[string]bool{
		"1.3.6.1.4.1.25623.1.0.108560": true,
	}
	tmpDuplicateRemoval := make(map[string]map[string]string)

	var filteredAlerts []Alert
	for _, alert := range alerts {
		if alert.Severity == "" {
			continue
		}

		severityText := alert.Severity
		if severityText == "Log" {
			alert.Severity = "Info"
		}

		oid := alert.NVTOID
		// skip if OID is blacklisted
		if blacklistedOIDList[oid] {
			continue
		}

		// skip if OID is already in the list
		if tmpDuplicateRemoval[oid] != nil {
			if tmpDuplicateRemoval[oid]["alert_title"] == alert.NVTName && tmpDuplicateRemoval[oid]["port"] == alert.Port {
				continue
			}
		}

		filteredAlerts = append(filteredAlerts, alert)

		// track duplicates
		if tmpDuplicateRemoval[oid] == nil {
			tmpDuplicateRemoval[oid] = make(map[string]string)
		}
		tmpDuplicateRemoval[oid]["alert_title"] = alert.NVTName
		tmpDuplicateRemoval[oid]["port"] = alert.Port

	}

	return filteredAlerts
}

func replaceStr(fileContent []byte) []byte {
	// write the file
	updatedContent := bytes.ReplaceAll(fileContent, []byte("OpenVAS"), []byte("IVA"))
	updatedContent = bytes.ReplaceAll(updatedContent, []byte("openvas"), []byte("IVA"))
	updatedContent = bytes.ReplaceAll(updatedContent, []byte("Greenbone"), []byte("IVA"))
	updatedContent = bytes.ReplaceAll(updatedContent, []byte("GreenBone"), []byte("IVA"))
	updatedContent = bytes.ReplaceAll(updatedContent, []byte("greenbone"), []byte("IVA"))
	updatedContent = bytes.ReplaceAll(updatedContent, []byte("green-bone"), []byte("IVA"))
	updatedContent = bytes.ReplaceAll(updatedContent, []byte("greenBone"), []byte("IVA"))

	return updatedContent
}
