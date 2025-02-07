package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/CSPF-Founder/iva/scanner/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ScanResultRepo struct {
	collection *mongo.Collection
}

func NewScanResultRepo(db *mongo.Database) ScanResultRepo {
	return ScanResultRepo{collection: db.Collection("scan_results")}
}

func (c *ScanResultRepo) AddList(
	ctx context.Context,
	records []models.ScanResult,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if records == nil {
		return 0, errors.New("records cannot be empty")
	}

	var interfaceRecords []any
	for _, r := range records {
		interfaceRecords = append(interfaceRecords, r)
	}

	insertedResult, err := c.collection.InsertMany(ctx, interfaceRecords)
	if err != nil {
		return 0, err
	}
	return len(insertedResult.InsertedIDs), nil
}

func (c *ScanResultRepo) Exists(ctx context.Context, query map[string]any) bool {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if id, ok := query["_id"]; ok {
		// Checking  if _id is already an ObjectID
		if oid, isObjectID := id.(primitive.ObjectID); isObjectID {

			query["_id"] = oid // If _id is already an ObjectID, directly use it in the query

		} else if idStr, isString := id.(string); isString {
			// If _id is a string, try converting it to ObjectID
			correctObjectID, err := primitive.ObjectIDFromHex(idStr)
			if err != nil {
				// Handle the error if conversion fails
				return false
			}
			query["_id"] = correctObjectID
		} else {
			return false
		}
	}
	result := c.collection.FindOne(ctx, query)
	return result.Err() == nil
}
