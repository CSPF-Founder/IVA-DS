package helpers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/CSPF-Founder/iva/panel/internal/sessions"
	"github.com/CSPF-Founder/iva/panel/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BaseData struct {
	Version                string
	ProductTitle           string
	CopyrightFooterCompany string
	Title                  string
	Flashes                []sessions.SessionFlash
	User                   models.User
	CSRFToken              string
	CSRFName               string
	CurrentYear            int
	PreviousPage           string
}

// Custom function to format tim.Time
func FormatNormalDate(date time.Time, layout string) string {
	return date.Format(layout)
}

// Custom function to format Mongo date
func FormatDate(date primitive.DateTime, layout string) string {
	return date.Time().Format(layout)
}

// Custom function to format Mongo date
func GetObjectIDString(id primitive.ObjectID) string {
	return id.Hex()
}

func ConvertJSONToString(data any) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	return string(jsonBytes)
}
