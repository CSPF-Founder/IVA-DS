package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	ctx "github.com/CSPF-Founder/iva/panel/context"
	"github.com/CSPF-Founder/iva/panel/enums"
	"github.com/CSPF-Founder/iva/panel/internal/repositories/datarepos"
	"github.com/CSPF-Founder/iva/panel/internal/services"
	"github.com/CSPF-Founder/iva/panel/internal/validator"
	mid "github.com/CSPF-Founder/iva/panel/middlewares"
	"github.com/CSPF-Founder/iva/panel/models"
	"github.com/CSPF-Founder/iva/panel/models/datamodels"
	"github.com/CSPF-Founder/iva/panel/utils"
	"github.com/CSPF-Founder/iva/panel/views"
	"github.com/CSPF-Founder/iva/panel/views/pages/targetpages"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type targetController struct {
	*App
	dataDB         *mongo.Database
	targetRepo     datarepos.TargetRepository
	scanResultRepo datarepos.ScanResultRepository
}

func newTargetController(
	app *App,
	dataDB *mongo.Database,
	targetRepo datarepos.TargetRepository,
	scanResultRepo datarepos.ScanResultRepository,
) *targetController {
	return &targetController{
		App:            app,
		dataDB:         dataDB,
		targetRepo:     targetRepo,
		scanResultRepo: scanResultRepo,
	}
}

func (c *targetController) registerRoutes() http.Handler {
	router := chi.NewRouter()

	// Authenticated Routes
	router.Group(func(r chi.Router) {
		r.Use(mid.RequireLogin)

		r.Get("/list", c.showScans)                       // List Targets table
		r.Post("/list", c.fetchScanData)                  // Get All Targets List
		r.Get("/add", c.displayAdd)                       // Display add Target form
		r.Post("/add", c.addHandler)                      // Handle add target form
		r.Post("/rescan", c.rescanHandler)                // Handle Rescan target form
		r.Post("/mark-as-main", c.markMainAlertHandler)   // Handle mark-as-main alert
		r.Post("/check-multi-status", c.checkMultiStatus) // Check multiple target status

		// eg: DELETE /scans/1
		r.Route("/{targetID:[0-9a-fA-F]{24}}", func(r chi.Router) {
			r.Delete("/", c.delete) // Handle Delete action
			r.Get("/report", c.downloadReport)
			r.Post("/alerts/{alertID}/status", c.updateAlertStatusHandler) // Handle mark-as-main alert

			scanResult := newScanResultController(c.App,
				c.scanResultRepo,
				c.targetRepo,
			)
			dsResult := newDSResultController(
				c.App,
				c.dataDB,
				c.targetRepo,
			)
			r.Mount("/scan-results", scanResult.registerRoutes())
			r.Mount("/ds-results", dsResult.registerRoutes())

		})

	})

	return router
}

// Show Add Target Page
func (c *targetController) displayAdd(w http.ResponseWriter, r *http.Request) {
	isDS := r.URL.Query().Get("is_ds") == "true"

	templateData := views.NewBaseData(c.config, c.session, r)
	title := "Add Scan"
	if isDS {
		title = "Add Differential Scan"
	}
	templateData.Title = title
	if err := views.RenderTempl(targetpages.AddTarget(title, isDS), templateData, w, r); err != nil {
		c.logger.Error("Error rendering template", err)
	}
}

// Add Target Handler
func (c *targetController) addHandler(w http.ResponseWriter, r *http.Request) {
	targetAddress := r.PostFormValue("target_address")
	targetAddress = strings.TrimSpace(targetAddress)

	isDsValue := r.PostFormValue("is_ds")

	targetAddress = strings.TrimSpace(targetAddress)

	targetType := enums.ParseTargetType(targetAddress)
	if targetType == "" || targetType == enums.TargetTypeInvalid {
		c.SendJSONError(w, "Invalid Target")
		return
	}

	isDs := false
	// Mark isDs true if value exists
	if isDsValue != "" {
		isDs = true
	}

	user := ctx.Get(r, "user").(models.User)
	target := &datamodels.Target{
		TargetAddress: targetAddress,
		Flag:          enums.ScanFlagWaitingToStart,
		ScanStatus:    enums.TargetStatusYetToStart,
		TargetType:    targetType,
		CustomerName:  user.Username,
		IsDS:          isDs,
	}

	err := c.targetRepo.SaveTarget(r.Context(), target)
	if err != nil {
		c.SendJSONError(w, "Unable to add the scan")
		return
	}
	c.SendJSONSuccess(w, "Successfully added the scan")
}

// Update Target Status Rescan
func (c *targetController) rescanHandler(w http.ResponseWriter, r *http.Request) {
	targetId := r.PostFormValue("target_id")

	user := ctx.Get(r, "user").(models.User)
	target, er := c.targetRepo.ByIdAndCustomerUsername(r.Context(), targetId, user.Username)
	if er != nil {
		c.flashAndGoBack(w, r, enums.FlashDanger, "Target Not Found")
		return
	}

	err := c.targetRepo.ResetScan(r.Context(), target)
	if err != nil {
		c.flashAndGoBack(w, r, enums.FlashWarning, "Unable to Rescan target, please try again later!")
		return
	}
	c.flash(w, r, enums.FlashSuccess, "Target submitted for Rescan", false)
	http.Redirect(w, r, "/targets/list?is_ds=true", http.StatusSeeOther)
}

// Show Target List
func (c *targetController) showScans(w http.ResponseWriter, r *http.Request) {

	isDS := r.URL.Query().Get("is_ds") == "true"
	templateData := views.NewBaseData(c.config, c.session, r)
	title := "View Scans"
	if isDS {
		title = "View Differential Scans"
	}

	templateData.Title = title
	if err := views.RenderTempl(targetpages.ListTarget(title, isDS), templateData, w, r); err != nil {
		c.logger.Error("Error rendering template", err)
	}
}

// Fetch All Target in json for ajax request
func (c *targetController) fetchScanData(w http.ResponseWriter, r *http.Request) {
	//Get is_ds from query param
	isDS := false
	if r.FormValue("is_ds") == "true" {
		isDS = true
	}

	defer func() {
		if r := recover(); r != nil {
			http.Error(w, "Unable to fetch the data", http.StatusUnprocessableEntity)
		}
	}()

	// Parse request parameters
	draw, err := strconv.Atoi(r.FormValue("draw"))
	if err != nil {
		http.Error(w, "Invalid draw parameter", http.StatusBadRequest)
		return
	}

	start, err := strconv.ParseInt(r.FormValue("start"), 10, 64)
	if err != nil || start < 0 {
		start = 0
	}

	length, err := strconv.ParseInt(r.FormValue("length"), 10, 64)
	if err != nil || length < 0 {
		length = 0
	}

	username := ctx.Get(r, "user").(models.User).Username // Fetch username from Auth::user()->getUsername()
	totalData, err := c.targetRepo.CountByCustomerUsername(r.Context(), username, isDS)
	if err != nil {
		c.logger.Error("Error fetching the data: ", err)
		http.Error(w, "Unable to fetch the data", http.StatusUnprocessableEntity)
		return
	}

	totalFiltered := totalData
	targetList, err := c.targetRepo.ListByCustomerUsername(r.Context(), username, isDS, start, length)
	if err != nil {
		c.logger.Error("Error fetching the data: ", err)
		http.Error(w, "Unable to fetch the data", http.StatusUnprocessableEntity)
		return
	}

	data := make([]map[string]any, 0)

	for _, target := range targetList {

		nestedData := make(map[string]any)
		nestedData["id"] = fmt.Sprintf("%v", target.ID.Hex())
		nestedData["target_address"] = validator.SanitizeXss(target.TargetAddress)

		if target.ScanStatus == enums.TargetStatusScanStarted {
			nestedData["scan_status_text"] = `<span class="spinner-border spinner-border-sm text-primary" aria-hidden="true"></span> <span role="status">Scanning...</span>`
		} else {
			nestedData["scan_status_text"] = target.GetScanStatusText()
		}

		nestedData["scan_status"] = target.ScanStatus
		scanStartedText := ""
		scanCompletedText := ""
		if target.ScanStartedTime == nil {
			scanStartedText = "-"
		} else {
			scanStartedText = target.ScanStartedTimeStr()
		}
		if target.ScanCompletedTime == nil {
			scanCompletedText = "-"
		} else {
			scanCompletedText = target.ScanCompletedTimeStr()
		}
		nestedData["scan_started_time"] = scanStartedText
		nestedData["scan_completed_time"] = scanCompletedText

		action := ""

		if canShowReportBtn(target) {
			action += fmt.Sprintf(`<a class="btn btn-sm btn-primary m-1 report-button" href="/targets/%v/report">Report</a>`, target.ID.Hex())
		} else {
			action += `<a class="btn btn-sm btn-dark m-1 disabled report-button" disabled href="#">Report</a>`
		}

		if canShowAlertsBtn(target) {
			if target.IsDS {
				action += fmt.Sprintf(`<a class="btn btn-sm btn-primary m-1 alerts-button" href="/targets/%v/ds-results">Alerts</a>`, target.ID.Hex())
			} else {
				action += fmt.Sprintf(`<a class="btn btn-sm btn-primary m-1 alerts-button" href="/targets/%v/scan-results">Alerts</a>`, target.ID.Hex())
			}
		} else {
			action += `<a class="btn btn-sm btn-dark m-1 disabled alerts-button" disabled href="#">Alerts</a>`
		}

		if canShowDeleteBtn(target) {
			action += fmt.Sprintf(`<button data-id="%v" class="btn btn-sm btn-danger text-white m-1 delete-target">Delete</button>`, target.ID.Hex())
		} else {
			action += fmt.Sprintf(`<button data-id="%v" class="btn btn-sm btn-dark text-white m-1 delete-target disabled" disabled>Delete</button>`, target.ID.Hex())
		}

		nestedData["action"] = action

		data = append(data, nestedData)
	}

	jsonOutput := map[string]any{
		"draw":            draw,
		"recordsTotal":    totalData,
		"recordsFiltered": totalFiltered,
		"records":         data,
	}

	w.Header().Set("Content-Type", "application/json")
	c.App.JSONResponse(w, jsonOutput, 200)
}

// conditions to meet to show the alerts button
// 1. If the scan status is ReportGenerated
// (or)
//  2. If it is DS scan and there are previous scans available
//     and the scan status is ScanFailed or Unreachable
func canShowAlertsBtn(target datamodels.Target) bool {
	return target.ScanStatus == enums.TargetStatusReportGenerated || (len(target.Scans) > 0 && (target.ScanStatus == enums.TargetStatusScanFailed || target.ScanStatus == enums.TargetStatusUnreachable))
}

func canShowReportBtn(target datamodels.Target) bool {
	return target.ScanStatus == enums.TargetStatusReportGenerated
}

func canShowDeleteBtn(target datamodels.Target) bool {
	return target.ScanStatus != enums.TargetStatusScanStarted
}

// Function to hande delete
func (c *targetController) delete(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")

	user := ctx.Get(r, "user").(models.User)

	// Fetching the target from the database
	target, err := c.targetRepo.ByIdAndCustomerUsername(r.Context(), targetID, user.Username)
	if err != nil {
		c.SendJSONError(w, "Invalid request")
		return
	}

	if target.IsDS {
		_, err := c.scanResultRepo.DeleteScanResultByTargetID(r.Context(), target.ID)
		if err != nil {
			c.logger.Error("error deleting scan results", err)
			c.SendJSONError(w, "Unable to delete the scan", http.StatusUnprocessableEntity)
			return
		}
	}

	err = services.DeleteTarget(r.Context(), *target, c.targetRepo)
	if err != nil {
		c.logger.Error("error deleting target", err)
		c.SendJSONError(w, "Unable to delete the scan", http.StatusUnprocessableEntity)
	} else {
		c.SendJSONSuccess(w, "Successfully deleted the scan", http.StatusOK)

	}
}

func (c *targetController) updateAlertStatusHandler(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	user := ctx.Get(r, "user").(models.User)

	// Fetching the target from the database
	target, err := c.targetRepo.ByIdAndCustomerUsername(r.Context(), targetID, user.Username)
	if err != nil {
		c.SendJSONError(w, "Invalid request")
		return
	}

	alertID := chi.URLParam(r, "alertID")
	alertObjectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		c.SendJSONError(w, "Invalid Alert Id")
		return
	}

	requiredParams := []string{"flag"}
	if !utils.CheckAllParamsExist(r, requiredParams) {
		c.flashAndGoBack(w, r, enums.FlashDanger, "Invalid Request")
		return
	}

	flag, err := enums.AlertStatusMap.ByIndex(r.FormValue("flag"))
	if err != nil {
		c.flashAndGoBack(w, r, enums.FlashDanger, "Invalid Alert Status")
		return
	}

	if !(flag == enums.AlertFalsePositive || flag == enums.AlertIgnored) {
		c.flashAndGoBack(w, r, enums.FlashDanger, "Invalid Alert Status")
		return
	}

	dsResultRepo := datarepos.NewDSResultRepo(c.dataDB, target.ID)
	_, err = dsResultRepo.ByID(r.Context(), alertObjectID)
	if err != nil {
		c.logger.Error("Error fetching alert: ", err)
		c.flashAndGoBack(w, r, enums.FlashDanger, "Alert Not Found")
		return
	}

	_, u_err := c.scanResultRepo.UpdateAlertStatus(r.Context(), alertObjectID, target.ID, flag)
	if u_err != nil {
		c.flashAndGoBack(w, r, enums.FlashDanger, "Error Updating Alert Status")
		return
	}
	c.flashAndGoBack(w, r, enums.FlashSuccess, "Alert Status Updated Succesfully")
}

func (c *targetController) markMainAlertHandler(w http.ResponseWriter, r *http.Request) {
	targetID := r.PostFormValue("target_id")

	user := ctx.Get(r, "user").(models.User)

	// Fetching the target from the database
	target, err := c.targetRepo.ByIdAndCustomerUsername(r.Context(), targetID, user.Username)
	if err != nil {
		c.flashAndGoBack(w, r, enums.FlashDanger, "Target Not Found!")
		return
	}

	if len(target.Scans) <= 0 {
		c.flashAndGoBack(w, r, enums.FlashDanger, "No Scan Numbers Available!")
		return
	}
	//Get Last/Latest scan number from the scan_numbers list
	scanNumber := target.Scans[len(target.Scans)-1].ScanNumber

	_, err = c.scanResultRepo.DeleteScanByID(r.Context(), target.ID, scanNumber)
	if err != nil {
		c.flashAndGoBack(w, r, enums.FlashDanger, "Unable to make this as main")
		return
	}

	_, err = c.targetRepo.RemoveScanNumbersFromTarget(r.Context(), target.ID)
	if err != nil {
		c.flashAndGoBack(w, r, enums.FlashDanger, "Unable to make this as main")
		return
	}

	c.flashAndGoBack(w, r, enums.FlashSuccess, "Successfully Make this as main")
}

// Function to Download Report
func (c *targetController) downloadReport(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")

	user := ctx.Get(r, "user").(models.User)
	target, err := c.targetRepo.ByIdAndCustomerUsername(r.Context(), targetID, user.Username)
	if err != nil {
		c.flashAndGoBack(w, r, enums.FlashWarning, "Invalid Request")
		return
	}

	err = services.DownloadReport(*target, w, r)
	if err != nil {
		c.logger.Error("Error downloading report: ", err)
		c.flashAndGoBack(w, r, enums.FlashWarning, "Unable to download the report")
	}
}

// Check Scan Status Handler Interval
func (c *targetController) checkMultiStatus(w http.ResponseWriter, r *http.Request) {
	requiredParams := []string{"target_ids[]"}

	if !utils.CheckAllParamsExist(r, requiredParams) {
		c.SendJSONError(w, "Please fill all the inputs", http.StatusUnprocessableEntity)
		return
	}

	targetIds := r.PostForm["target_ids[]"]

	user := ctx.Get(r, "user").(models.User)
	targetList, err := c.targetRepo.ListByIdsAndCustomerUsername(r.Context(), targetIds, user.Username)
	if err != nil {
		c.SendJSONError(w, "Invalid Request", http.StatusUnprocessableEntity)
		return
	}

	if targetList == nil {
		c.SendJSONError(w, "Invalid Request", http.StatusUnprocessableEntity)
		return
	}

	var entries []map[string]any

	for _, target := range targetList {
		entry := map[string]any{
			"id":                  target.ID.Hex(),
			"scan_status":         target.ScanStatus,
			"scan_status_text":    target.GetScanStatusText(),
			"scan_started_time":   target.ScanStartedTimeStr(),
			"scan_completed_time": target.ScanCompletedTimeStr(),
			"show_report_btn":     canShowReportBtn(target),
			"show_alerts_btn":     canShowAlertsBtn(target),
			"show_delete_btn":     canShowDeleteBtn(target),
		}
		entries = append(entries, entry)
	}

	c.JSONResponse(w, map[string]any{"data": entries}, http.StatusOK)

}
