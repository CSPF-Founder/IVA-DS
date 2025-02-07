package web

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/CSPF-Founder/iva/scanner/enums"
	"github.com/CSPF-Founder/iva/scanner/internal/repositories"
	"github.com/CSPF-Founder/iva/scanner/logger"
	"github.com/CSPF-Founder/iva/scanner/models"
	"github.com/CSPF-Founder/libs/zapapi/zap"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DSHandler struct {
	db      *repositories.Repository
	logger  *logger.Logger
	Target  models.Target
	Results []models.ScanResult
}

func NewDSHandler(
	db *repositories.Repository,
	logger *logger.Logger,
	target models.Target,
) *DSHandler {
	return &DSHandler{
		db:     db,
		logger: logger,
		Target: target,
	}
}

type DSCompareFields struct {
	VulnerabilityTitle string
	URL                string
	Param              string
}

// Handle handles the scan results
func (h *DSHandler) handle(ctx context.Context, results []zap.Alert) error {
	firstScan := len(h.Target.Scans) == 0

	if firstScan {
		err := h.handleFirstTime(ctx, results)
		if err != nil {
			return err
		}
		return nil
	}
	err := h.handleSubsequent(ctx, results)
	if err != nil {
		return err
	}
	return nil
}

// handleFirstTime handles the first scan for a target
func (h *DSHandler) handleFirstTime(ctx context.Context, results []zap.Alert) error {
	scanNum := 1
	err := h.updateScansInfo(ctx, scanNum)
	if err != nil {
		return fmt.Errorf("failed to update scan info %w", err)
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	records := make([]models.ScanResult, 0, len(results))
	for _, entry := range results {
		record, err := parseRecord(entry, h.Target)
		if err != nil {
			h.logger.Error("Error converting entry to record", err)
			continue
		}

		record.FoundDate = &now
		record.AlertStatus = enums.AlertStatusUnfixed
		record.ScanNumbers = append(record.ScanNumbers, scanNum)

		records = append(records, *record)
	}

	if len(records) == 0 {
		h.logger.Info("No records to add to the database")
		return nil
	}

	_, err = h.db.DSResult.AddList(ctx, records, h.Target.ID)
	if err != nil {
		return err
	}
	return nil
}

type processedData struct {
	newRecords     []models.ScanResult
	unFixedEntries map[primitive.ObjectID]bool
	fixedEntries   []primitive.ObjectID
}

func getFixedEntries(
	previousResults []models.ScanResult,
	unFixedEntries map[primitive.ObjectID]bool,
) []primitive.ObjectID {
	// The alert exist in the previous scan but not exist in the current scan,
	//	that means the alert is fixed
	fixedEntries := []primitive.ObjectID{}
	for _, item := range previousResults {
		if item.AlertStatus == enums.AlertStatusIgnored || item.AlertStatus == enums.AlertStatusFP {
			// if the alert is ignored or fp, then it should not be marked as fixed
			continue
		}

		if !unFixedEntries[item.ID] {
			fixedEntries = append(fixedEntries, item.ID)
		}
	}

	return fixedEntries
}

// processScanResults processes the scan results
func (h *DSHandler) processScanResults(
	results []zap.Alert,
	previousResults []models.ScanResult,
	scanNum int,
) (*processedData, error) {
	prevMap := convertPrevResultToMap(previousResults)
	unFixedEntries := map[primitive.ObjectID]bool{}
	newRecords := make([]models.ScanResult, 0, len(results))
	dupMap := make(mappedResults)

	for _, entry := range results {
		record, err := parseRecord(entry, h.Target)
		if err != nil {
			h.logger.Error("Error converting entry to record", err)
			continue
		}

		cmpKey := prepDSCmpFields(record.VulnerabilityTitle, record.WSData.URL, record.WSData.Param)

		if _, ok := dupMap[cmpKey]; ok {
			continue
		}

		dupMap[cmpKey] = mappedResultData{
			// No need to fill the fields, since we are not going to use it
		}

		// if the current alert in previous alert, it means that particular alert is not fixed
		if prev, ok := prevMap[cmpKey]; ok {
			if prev.AlertStatus != enums.AlertStatusIgnored && prev.AlertStatus != enums.AlertStatusFP {
				// only add to unfixed entries if the alert is not ignored or fp
				unFixedEntries[prev.ID] = true
			}
			continue
		}

		now := primitive.NewDateTimeFromTime(time.Now())
		// If the current alert not in previous alert, it means It is New alert
		record.FoundDate = &now
		record.AlertStatus = enums.AlertStatusUnfixed
		record.ScanNumbers = append(record.ScanNumbers, scanNum)
		newRecords = append(newRecords, *record)
	}

	return &processedData{
		newRecords:     newRecords,
		unFixedEntries: unFixedEntries,
		fixedEntries:   getFixedEntries(previousResults, unFixedEntries),
	}, nil
}

// handleSubsequent handles the subsequent scans for a target
func (h *DSHandler) handleSubsequent(ctx context.Context, results []zap.Alert) error {
	scanNum := len(h.Target.Scans) + 1

	err := h.updateScansInfo(ctx, scanNum)
	if err != nil {
		return err
	}

	previousResults, err := h.db.DSResult.GetWebScanAlertsByTarget(ctx, h.Target.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch previous results %w", err)
	}

	procData, err := h.processScanResults(results, previousResults, scanNum)
	if err != nil {
		return fmt.Errorf("failed to process scan results %w", err)
	}

	if len(procData.unFixedEntries) > 0 {
		err := h.updateUnfixedEntries(ctx, scanNum, procData.unFixedEntries)
		if err != nil {
			return err
		}
	}

	if len(procData.fixedEntries) > 0 {
		err := h.updateFixedEntries(ctx, procData.fixedEntries)
		if err != nil {
			return err
		}
	}

	if len(procData.newRecords) > 0 {
		_, err := h.db.DSResult.AddList(ctx, procData.newRecords, h.Target.ID)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

// updateScansInfo update the `scans` entry
func (h *DSHandler) updateScansInfo(
	ctx context.Context,
	scanNum int,
) error {
	scanInfo := models.ScanInfo{
		ScanNumber: scanNum,
		ScanDate:   time.Now(),
	}

	h.Target.Scans = append(h.Target.Scans, scanInfo)
	_, err := h.db.Target.UpdateScanInfo(ctx, h.Target.ID, scanInfo)
	if err != nil {
		return err
	}
	return nil
}

// updateExistingScanResult updates the existing scan results
func (h *DSHandler) updateUnfixedEntries(
	ctx context.Context,
	scanNumber int,
	ids map[primitive.ObjectID]bool,
) error {
	var bulkWrites []mongo.WriteModel
	for id := range ids {
		filter := bson.M{"_id": id}
		update := bson.M{"$set": bson.M{
			"alert_status": enums.AlertStatusUnfixed,
		}, "$push": bson.M{
			"scan_numbers": scanNumber,
		}}
		bulkWrites = append(bulkWrites, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}

	if len(bulkWrites) > 0 {
		err := h.db.DSResult.BulkWrite(ctx, h.Target.ID, bulkWrites)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateFixedEntries updates the fixed entries
func (h *DSHandler) updateFixedEntries(
	ctx context.Context,
	ids []primitive.ObjectID,
) error {
	var bulkWrites []mongo.WriteModel

	now := time.Now()

	for _, id := range ids {
		filter := bson.M{"_id": id}
		update := bson.M{"$set": bson.M{"alert_status": enums.AlertStatusFixed, "fixed_date": now}}
		bulkWrites = append(bulkWrites, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}

	if len(bulkWrites) > 0 {
		err := h.db.DSResult.BulkWrite(ctx, h.Target.ID, bulkWrites)
		if err != nil {
			return err
		}
	}
	return nil
}

type mappedResultData struct {
	ID          primitive.ObjectID
	AlertStatus enums.AlertStatus
}
type mappedResults map[DSCompareFields]mappedResultData

// convertPrevResultToMap converts the previous results to a map
func convertPrevResultToMap(entries []models.ScanResult) mappedResults {
	resultMap := make(mappedResults)
	for _, entry := range entries {
		var url string
		var param string
		if entry.WSData != nil {
			url = entry.WSData.URL
			param = entry.WSData.Param
		}
		key := prepDSCmpFields(
			entry.VulnerabilityTitle,
			url,
			param,
		)
		resultMap[key] = mappedResultData{
			ID:          entry.ID,
			AlertStatus: entry.AlertStatus,
		}
	}
	return resultMap
}

// prepDSCmpFields converts the fields to comparable fields
func prepDSCmpFields(title string, url string, param string) DSCompareFields {
	title = strings.TrimSpace(title)
	title = strings.ToLower(title)

	url = strings.TrimSpace(url)
	url = strings.ToLower(url)

	param = strings.TrimSpace(param)
	param = strings.ToLower(param)

	return DSCompareFields{
		VulnerabilityTitle: title,
		URL:                url,
		Param:              param,
	}
}
