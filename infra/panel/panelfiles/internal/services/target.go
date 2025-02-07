package services

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/CSPF-Founder/iva/panel/internal/repositories/datarepos"
	"github.com/CSPF-Founder/iva/panel/models/datamodels"
)

func DownloadReport(target datamodels.Target, w http.ResponseWriter, r *http.Request) error {
	reportPath, err := target.GetReportPath()
	if err != nil {
		return err
	}
	reportID := target.ID.Hex()

	_, err = os.Stat(reportPath)

	if err == nil {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("filename=\"IVA_report_%s.docx\"", reportID))
		http.ServeFile(w, r, reportPath)
		return nil
	} else {
		return err
	}
}

func DeleteTarget(ctx context.Context, target datamodels.Target, targetRepo datarepos.TargetRepository) error {
	if !target.CanDelete() {
		return fmt.Errorf("target cannot be deleted")
	}

	reportDir, err := target.GetReportDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(reportDir); err == nil {
		// Delete report folder
		os.RemoveAll(reportDir)
	}

	// Delete target
	deletedCount, err := targetRepo.DeleteTargetByID(ctx, target.ID)
	if err != nil {
		return err
	}

	if deletedCount > 0 {
		return nil
	} else {
		return fmt.Errorf("target not found")
	}
}
