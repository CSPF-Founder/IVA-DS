package controllers

import (
	"fmt"
	"net/http"

	ctx "github.com/CSPF-Founder/iva/panel/context"
	"github.com/CSPF-Founder/iva/panel/enums"
	"github.com/CSPF-Founder/iva/panel/internal/repositories/datarepos"
	"github.com/CSPF-Founder/iva/panel/internal/services"
	mid "github.com/CSPF-Founder/iva/panel/middlewares"
	"github.com/CSPF-Founder/iva/panel/models"
	"github.com/CSPF-Founder/iva/panel/utils/iputils"
	"github.com/CSPF-Founder/iva/panel/views"
	"github.com/CSPF-Founder/iva/panel/views/pages/alertspages"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type dsResultController struct {
	*App
	dataDB     *mongo.Database
	targetRepo datarepos.TargetRepository
}

func newDSResultController(app *App,
	dataDB *mongo.Database,
	targetRepo datarepos.TargetRepository,
) *dsResultController {
	return &dsResultController{
		App:        app,
		dataDB:     dataDB,
		targetRepo: targetRepo,
	}
}

func (c *dsResultController) registerRoutes() http.Handler {
	router := chi.NewRouter()

	// Authenticated Routes
	router.Group(func(r chi.Router) {
		r.Use(mid.RequireLogin)

		r.Get("/", c.list)         // List all Scan Results
		r.Get("/export", c.export) // Export Alerts
	})

	return router
}

func (c *dsResultController) list(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")

	user := ctx.Get(r, "user").(models.User)
	target, err := c.targetRepo.ByIdAndCustomerUsername(r.Context(), targetID, user.Username)
	if err != nil {
		c.logger.Error("Error getting target", err)
		c.flashAndGoBack(w, r, enums.FlashWarning, "Target Not Found")
		return
	}

	dsResultRepo := datarepos.NewDSResultRepo(c.dataDB, target.ID)

	unfixedAlerts, err := dsResultRepo.ListByAlertStatus(r.Context(), enums.AlertUnfixed)
	if err != nil {
		c.logger.Error("Error getting unfixed alerts", err)
		c.flashAndGoBack(w, r, enums.FlashDanger, "Error getting unfixed alerts")
		return
	}

	fixedAlerts, err := dsResultRepo.ListByAlertStatus(r.Context(), enums.AlertFixed)
	if err != nil {
		c.logger.Error("Error getting fixed alerts", err)
		c.flashAndGoBack(w, r, enums.FlashDanger, "Error getting fixed alerts")
		return
	}

	ignoredAlerts, err := dsResultRepo.ListByAlertStatus(r.Context(), enums.AlertIgnored)
	if err != nil {
		c.logger.Error("Error getting ignored alerts", err)
		c.flashAndGoBack(w, r, enums.FlashDanger, "Error getting ignored alerts")
		return
	}

	fpAlerts, err := dsResultRepo.ListByAlertStatus(r.Context(), enums.AlertFalsePositive)
	if err != nil {
		c.logger.Error("Error getting false positive alerts", err)
		c.flashAndGoBack(w, r, enums.FlashDanger, "Error getting false positive alerts")
		return
	}

	if target.ScanStatus == enums.TargetStatusUnreachable {
		if target.TargetType == enums.TargetTypeURL {
			// URL
			c.flash(w, r, enums.FlashDanger, "Target is not reachable at current scan. The alerts status is not updated.", false)
			c.flash(w, r, enums.FlashWarning, "Please check the target status and rescan", false)
		} else {
			// IP/IP range
			c.flash(w, r, enums.FlashDanger, "No open ports found. The alerts status is not updated.", false)
			c.flash(w, r, enums.FlashWarning, "Please check the target status and rescan", false)
		}
	} else if len(unfixedAlerts) > 0 {
		c.flash(w, r, enums.FlashWarning, "Ensure to mark False Positive and Ignore", false)
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

	for _, result := range unfixedAlerts {
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
		CanRescan:                           target.CanRescan(),
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

		alertsData := alertspages.DSMultiTargetData{
			Unfixed: groupResultsByIP(unfixedAlerts),
			Fixed:   groupResultsByIP(fixedAlerts),
			Ignored: groupResultsByIP(ignoredAlerts),
			FP:      groupResultsByIP(fpAlerts),
		}

		component = alertspages.DSListIPRangeResults(commonData, alertsData)

	} else {
		commonData.TotalTargets = 1
		alertsData := alertspages.DSSingleTargetData{
			Unfixed: unfixedAlerts,
			Fixed:   fixedAlerts,
			Ignored: ignoredAlerts,
			FP:      fpAlerts,
		}
		component = alertspages.DSListResults(commonData, alertsData)
	}

	if err := views.RenderTempl(component, templateData, w, r); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

func (c *dsResultController) export(w http.ResponseWriter, r *http.Request) {
	target_id := chi.URLParam(r, "targetID")
	user := ctx.Get(r, "user").(models.User)

	target, err := c.targetRepo.ByIdAndCustomerUsername(r.Context(), target_id, user.Username)
	if err != nil {
		c.flashAndGoBack(w, r, enums.FlashWarning, "No Target Found")
		return
	}

	dsResultRepo := datarepos.NewDSResultRepo(c.dataDB, target.ID)

	bugList, err := dsResultRepo.GetDetailAlertsByTarget(r.Context())

	if err != nil {
		c.flashAndGoBack(w, r, enums.FlashWarning, "No Alerts Found")
		return
	}

	if len(bugList) > 0 {
		f := excelize.NewFile()

		// Set column widths
		f.SetColWidth("Sheet1", "A", "A", 42)
		f.SetColWidth("Sheet1", "B", "B", 40)
		f.SetColWidth("Sheet1", "C", "C", 15) // Auto size
		f.SetColWidth("Sheet1", "D", "D", 40)
		f.SetColWidth("Sheet1", "E", "E", 40)
		f.SetColWidth("Sheet1", "F", "F", 40)
		f.SetColWidth("Sheet1", "G", "G", 20) // Auto size
		f.SetColWidth("Sheet1", "H", "H", 20) // Auto size

		// Set header values
		headers := []string{
			"Url/IP/Application",
			"Alert",
			"Severity",
			"Details/Impact",
			"Replication/Proof",
			"Remediation",
			"Found Date",
			"Status",
		}
		for i, header := range headers {
			cell := fmt.Sprintf("%s1", string(rune(65+i)))
			f.SetCellValue("Sheet1", cell, header)
		}

		// Style header
		styleHeader, _ := f.NewStyle(`{
			"font": {"bold": true, "color": "#ffffff"},
			"fill": {"type": "pattern", "color": ["#FAA61A"], "pattern":1},
			"alignment": {"horizontal": "center", "vertical": "center"}
		}`)
		f.SetCellStyle("Sheet1", "A1", "H1", styleHeader)

		x := 2
		for _, bug := range bugList {
			formatedFoundDate := ""
			if bug.FoundDate != nil {
				formatedFoundDate = bug.FoundDate.Time().Format("02-01-2006")
			}
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", x), target.TargetAddress)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", x), bug.VulnerabilityTitle)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", x), enums.SeverityToString(bug.Severity))
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", x), bug.GetDetailsAndImpact())
			f.SetCellValue("Sheet1", fmt.Sprintf("E%d", x), bug.GetPOC())
			f.SetCellValue("Sheet1", fmt.Sprintf("F%d", x), bug.Remediation)
			f.SetCellValue("Sheet1", fmt.Sprintf("G%d", x), formatedFoundDate)
			f.SetCellValue("Sheet1", fmt.Sprintf("H%d", x), enums.AlertStatusToString(bug.AlertStatus))

			// Style severity
			severityColor := "#FF0000"
			switch bug.Severity {
			case enums.SeverityCritical:
				severityColor = "#E83123"
			case enums.SeverityHigh:
				severityColor = "#E77f34"
			case enums.SeverityMedium:
				severityColor = "#e6ac30"
			case enums.SeverityLow:
				severityColor = "#2fa84d"
			case enums.SeverityInfo:
				severityColor = "#0773b8"
			}
			styleSeverity, _ := f.NewStyle(`{
				"font": {"bold": true, "color": "#FFFFFF"},
				"fill": {"type": "pattern", "color": ["` + severityColor + `"], "pattern":1},
				"alignment":{"horizontal":"center","vertical":"center","wrap_text":true}
			}`)

			f.SetCellStyle("Sheet1", fmt.Sprintf("C%d", x), fmt.Sprintf("C%d", x), styleSeverity)

			// Style Row I to L
			styleRow, _ := f.NewStyle(`{
				"alignment": {"horizontal": "center", "vertical": "center","wrap_text":true}
			}`)
			f.SetCellStyle("Sheet1", fmt.Sprintf("G%d", x), fmt.Sprintf("H%d", x), styleRow)

			// Style Row
			styleRow2, _ := f.NewStyle(`{
				"alignment": {"wrap_text":true, "vertical": "center"}
			}`)
			f.SetCellStyle("Sheet1", fmt.Sprintf("A%d", x), fmt.Sprintf("B%d", x), styleRow2)
			f.SetCellStyle("Sheet1", fmt.Sprintf("D%d", x), fmt.Sprintf("F%d", x), styleRow2)

			// Status validation & style
			f.AddDataValidation("Sheet1", &excelize.DataValidation{
				Type:         "list",
				Formula1:     `"Unfixed,Fixed"`,
				ShowDropDown: true,
				Sqref:        fmt.Sprintf("H%d", x),
			})
			// f.SetCellDataValidation("Sheet1", fmt.Sprintf("L%d", x), "list", `"Unfixed,Fixed"`)
			// f.SetCellStyle("Sheet1", fmt.Sprintf("L%d", x), fmt.Sprintf("L%d", x), styleHeader)

			f.SetRowHeight("Sheet1", x, 120)
			x++
		}

		// Set sheet title
		f.SetSheetName("Sheet1", "Vulnerabilities")

		// Write to response
		w.Header().Set("Content-Disposition", "attachment; filename=Vulnerabilities.xlsx")
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		if err := f.Write(w); err != nil {
			fmt.Println(err)
		}
	} else {
		c.flashAndGoBack(w, r, enums.FlashWarning, "No Alerts Found")
		return
	}
}
