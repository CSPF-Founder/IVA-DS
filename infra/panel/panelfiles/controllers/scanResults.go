package controllers

import (
	"net/http"
	"sort"
	"strings"

	ctx "github.com/CSPF-Founder/iva/panel/context"
	"github.com/CSPF-Founder/iva/panel/enums"
	"github.com/CSPF-Founder/iva/panel/internal/repositories/datarepos"
	"github.com/CSPF-Founder/iva/panel/internal/services"
	mid "github.com/CSPF-Founder/iva/panel/middlewares"
	"github.com/CSPF-Founder/iva/panel/models"
	"github.com/CSPF-Founder/iva/panel/models/datamodels"
	"github.com/CSPF-Founder/iva/panel/utils/iputils"
	"github.com/CSPF-Founder/iva/panel/views"
	"github.com/CSPF-Founder/iva/panel/views/pages/alertspages"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

type scanResultController struct {
	*App
	scanResultRepo datarepos.ScanResultRepository
	targetRepo     datarepos.TargetRepository
}

func newScanResultController(app *App,
	scanResultRepo datarepos.ScanResultRepository,
	targetRepo datarepos.TargetRepository,
) *scanResultController {
	return &scanResultController{
		App:            app,
		scanResultRepo: scanResultRepo,
		targetRepo:     targetRepo,
	}
}

func (c *scanResultController) registerRoutes() http.Handler {
	router := chi.NewRouter()

	// Authenticated Routes
	router.Group(func(r chi.Router) {
		r.Use(mid.RequireLogin)

		r.Get("/", c.list) // List all Scan Results
	})

	return router
}

func (c *scanResultController) list(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")

	user := ctx.Get(r, "user").(models.User)
	target, err := c.targetRepo.ByIdAndCustomerUsername(r.Context(), targetID, user.Username)
	if err != nil {
		c.logger.Error("Error getting target", err)
		c.flashAndGoBack(w, r, enums.FlashWarning, "Target Not Found")
		return
	}

	c.listResults(w, r, target)
}

func (c *scanResultController) listResults(w http.ResponseWriter, r *http.Request, target *datamodels.Target) {
	scanResults, err := c.scanResultRepo.ListByTarget(r.Context(), target.ID)
	if err != nil {
		c.logger.Error("Error getting alerts", err)
		c.flashAndGoBack(w, r, enums.FlashDanger, "Error getting alerts")
		return
	}

	if len(scanResults) == 0 {
		c.flashAndGoBack(w, r, enums.FlashInfo, "No alerts found")
		return
	}

	vulnerabilityStats := alertspages.Vulnerability{
		Critical:          0,
		High:              0,
		Medium:            0,
		Low:               0,
		Info:              0,
		NoVulnerabilities: 1,
	}
	totalAlerts := 0

	for _, result := range scanResults {
		totalAlerts++
		switch result.Severity {
		case enums.SeverityCritical:
			vulnerabilityStats.Critical++
		case enums.SeverityHigh:
			vulnerabilityStats.High++
		case enums.SeverityMedium:
			vulnerabilityStats.Medium++
		case enums.SeverityLow:
			vulnerabilityStats.Low++
		case enums.SeverityInfo:
			vulnerabilityStats.Info++
		}
	}

	// Check if no vulnerabilities found
	totalAlertStats := vulnerabilityStats.Critical + vulnerabilityStats.High + vulnerabilityStats.Medium + vulnerabilityStats.Low + vulnerabilityStats.Info
	if totalAlertStats > 0 {
		vulnerabilityStats.NoVulnerabilities = 0
	}

	var component templ.Component

	templateData := views.NewBaseData(c.config, c.session, r)
	templateData.Title = "Scan Results"

	commonData := alertspages.CommonResultData{
		Target:                              target,
		ReportGenerated:                     enums.TargetStatus(enums.TargetStatusReportGenerated),
		VulnerabilityStats:                  vulnerabilityStats,
		OverallCVSSScore:                    target.OverallCVSSScore,
		TotalAlerts:                         totalAlerts,
		CVSSScoreByHost:                     target.CVSSScoreByHost,
		DefaultRemediation:                  services.DefaultRemediation,
		NumberOfTARowsForDefaultRemediation: services.NumberOfTARowsForDefaultRemediation,
	}

	if target.TargetType == enums.TargetTypeIPRange {
		ipCount, err := iputils.ConvertIPRangeToIPSize(target.TargetAddress)
		if err != nil {
			c.flashAndGoBack(w, r, enums.FlashDanger, "Invalid Target Address")
			return
		}
		commonData.TotalTargets = ipCount.Int64()

		records := groupResultsByIP(scanResults)
		component = alertspages.ListIPRangeResults(commonData, records)
	} else {
		commonData.TotalTargets = 1
		component = alertspages.ListResults(commonData, scanResults)
	}

	if err := views.RenderTempl(component, templateData, w, r); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

// hostToIp converts a host to an IP address
// if the host is http://192.168.56.1:80, it will return 192.168.56.1
// if the host is 192.168.56.1:80, it will return 192.168.56.1
// if the host is http://192.168.56.1/test it will return 192.168.56.1
// func hostToIp(entry string) string {
// 	u, err := url.Parse(entry)
// 	if err == nil && u.Host != "" {
// 		host, _, err := net.SplitHostPort(u.Host)
// 		if err == nil {
// 			return host
// 		}
// 		return u.Host
// 	}

// 	if strings.Contains(entry, "/") {
// 		entry = strings.Split(entry, "/")[0]
// 	}

//		host, _, err := net.SplitHostPort(entry)
//		if err != nil {
//			return entry
//		}
//		return host
//	}

func groupResultsByIP(scanResults []datamodels.ScanResult) map[string][]datamodels.ScanResult {
	records := make(map[string][]datamodels.ScanResult)
	if scanResults == nil {
		return records
	}

	for _, scanResult := range scanResults {
		if scanResult.NSData == nil {
			continue
		}

		ip := scanResult.NSData.IP
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		if _, ok := records[ip]; !ok {
			records[ip] = []datamodels.ScanResult{}
		}
		records[ip] = append(records[ip], scanResult)
	}

	// Sort the records by IP
	keys := make([]string, 0, len(records))
	for k := range records {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sortedRecords := make(map[string][]datamodels.ScanResult)
	for _, k := range keys {
		sortedRecords[k] = records[k]
	}

	return sortedRecords
}
