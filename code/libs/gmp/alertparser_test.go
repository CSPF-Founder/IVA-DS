package gmp

import (
	"testing"
)

func TestParseCSV(t *testing.T) {
	csvData := []byte(`IP,Hostname,Severity,NVT Name,NVT OID,Port,Port Protocol,CVSS,Summary,Vulnerability Insight,Impact,Solution,Other References,Affected Software/OS,CVEs,Specific Result
192.168.1.1,host1,High,Test NVT,1.3.6.1.4.1.25623.1.0.108560,80,TCP,7.5,Test Summary,Test Insight,Test Impact,Test Solution,Test References,Test OS,CVE-2021-1234,Test Result
192.168.1.2,host2,Log,Test NVT 2,1.3.6.1.4.1.25623.1.0.108561,443,TCP,5.0,Test Summary 2,Test Insight 2,Test Impact 2,Test Solution 2,,Test OS 2,CVE-2021-5678,Test Result 2`)

	alerts, err := parseCSVData(csvData, "test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	alert := alerts[0]
	if alert.IP != "192.168.1.2" {
		t.Errorf("Expected IP '192.168.1.2', got %s", alert.IP)
	}
	if alert.Severity != "Info" {
		t.Errorf("Expected Severity 'Info', got %s", alert.Severity)
	}
	if alert.NVTName != "Test NVT 2" {
		t.Errorf("Expected NVTName 'Test NVT 2', got %s", alert.NVTName)
	}
	if alert.NVTOID != "1.3.6.1.4.1.25623.1.0.108561" {
		t.Errorf("Expected NVTOID '1.3.6.1.4.1.25623.1.0.108561', got %s", alert.NVTOID)
	}

	if alert.OtherReferences != "" {
		t.Errorf("Expected empty OtherReferences, got %s", alert.OtherReferences)
	}
}
