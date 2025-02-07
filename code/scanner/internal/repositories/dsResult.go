package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CSPF-Founder/iva/scanner/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DSResultRepo struct {
	db *mongo.Database
}

func NewDSResultRepo(db *mongo.Database) DSResultRepo {
	return DSResultRepo{
		db: db,
	}
}

// DS Functions

func (ds *DSResultRepo) getCollection(
	targetID primitive.ObjectID,
) *mongo.Collection {

	collectionName := fmt.Sprintf("scan_results_%s", targetID.Hex())
	return ds.db.Collection(collectionName)
}

func (ds *DSResultRepo) AddList(
	ctx context.Context,
	records []models.ScanResult,
	targetID primitive.ObjectID,
) (int, error) {

	collection := ds.getCollection(targetID)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if records == nil {
		return 0, errors.New("records cannot be empty")
	}

	var interfaceRecords []any
	for _, r := range records {
		interfaceRecords = append(interfaceRecords, r)
	}

	insertedResult, err := collection.InsertMany(ctx, interfaceRecords)
	if err != nil {
		return 0, err
	}
	return len(insertedResult.InsertedIDs), nil
}

// Fetching the alerts based on targetID
func (ds *DSResultRepo) GetNetworkScanAlertsByTarget(
	ctx context.Context,
	targetID primitive.ObjectID,
) ([]models.ScanResult, error) {
	collection := ds.getCollection(targetID)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var scanResults []models.ScanResult
	if err = cursor.All(ctx, &scanResults); err != nil {
		return nil, err
	}

	return scanResults, nil
}
func (ds *DSResultRepo) GetWebScanAlertsByTarget(
	ctx context.Context,
	targetID primitive.ObjectID,
) ([]models.ScanResult, error) {
	collection := ds.getCollection(targetID)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var scanResults []models.ScanResult
	if err = cursor.All(ctx, &scanResults); err != nil {
		return nil, err
	}

	return scanResults, nil
}

// Updating alerts
func (ds *DSResultRepo) BulkWrite(
	ctx context.Context,
	targetID primitive.ObjectID,
	bulkWrites []mongo.WriteModel,
) error {
	collection := ds.getCollection(targetID)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	bulkOptions := options.BulkWrite().SetOrdered(false)
	_, err := collection.BulkWrite(ctx, bulkWrites, bulkOptions)
	return err
}
