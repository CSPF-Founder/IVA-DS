package datarepos

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/CSPF-Founder/iva/panel/enums"
	"github.com/CSPF-Founder/iva/panel/models/datamodels"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ScanResultRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewScanResultRepository(db *mongo.Database) ScanResultRepository {
	return ScanResultRepository{collection: db.Collection("scan_results"), db: db}
}

func (s *ScanResultRepository) getScanResultCollectionNameByTarget(targetId primitive.ObjectID) string {
	return fmt.Sprintf("scan_results_%s", targetId.Hex())
}

func (s *ScanResultRepository) getScanResultCollectionByTarget(ctx context.Context, targetId primitive.ObjectID) (*mongo.Collection, error) {
	collectionName := s.getScanResultCollectionNameByTarget(targetId)
	collectionExists, err := s.dbHasCollection(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	if !collectionExists {
		return nil, fmt.Errorf("Scan Results not exists!")
	}
	return s.db.Collection(collectionName), nil
}

func (s *ScanResultRepository) getCollectionByTarget(targetId primitive.ObjectID) *mongo.Collection {
	collectionName := s.getScanResultCollectionNameByTarget(targetId)
	return s.db.Collection(collectionName)
}

// Get Scan Results data by target_id,
func (s *ScanResultRepository) ListByTarget(ctx context.Context, target_id primitive.ObjectID) ([]datamodels.ScanResult, error) {
	var scanResults []datamodels.ScanResult

	options := options.Find()
	options.SetSort(bson.D{
		{Key: "severity", Value: 1},
	})

	filter := map[string]any{
		"target_id": target_id,
	}

	cursor, err := s.collection.Find(ctx, filter, options)
	if err != nil {
		return scanResults, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var scanResult datamodels.ScanResult
		err := cursor.Decode(&scanResult)
		if err != nil {
			return scanResults, err
		}
		scanResults = append(scanResults, scanResult)
	}

	return scanResults, nil
}

// Check if collection exists in db
func (s *ScanResultRepository) dbHasCollection(ctx context.Context, collectionName string) (bool, error) {
	coll, err := s.db.ListCollectionNames(ctx, bson.D{{Key: "name", Value: collectionName}})
	if err != nil {
		return false, fmt.Errorf("Error Getting Scan Results!")
	}
	return len(coll) == 1, nil
}

func (s *ScanResultRepository) ByID(ctx context.Context, id primitive.ObjectID) (datamodels.ScanResult, error) {
	var scanResult datamodels.ScanResult
	filter := map[string]any{
		"_id": id,
	}

	cursor := s.collection.FindOne(ctx, filter)
	if cursor.Err() != nil {
		return scanResult, cursor.Err()
	}
	err := cursor.Decode(&scanResult)
	if err != nil {
		return scanResult, err
	}
	return scanResult, nil
}

// Update Alert Status query handler
func (s *ScanResultRepository) UpdateAlertStatus(ctx context.Context, id primitive.ObjectID, targetId primitive.ObjectID, status enums.AlertStatus) (int64, error) {
	// Construct the filter for documents to delete
	// Define the filter (if you want to apply this operation to all documents, you can use bson.D{})
	filter := bson.D{
		{Key: "_id", Value: id},
	}
	// Update operation
	// Define the update to retain only the last element of the "scans" array

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "alert_status", Value: status}, // Set alert_status to 1
		}},
	}

	collection, err := s.getScanResultCollectionByTarget(ctx, targetId)

	if err != nil {
		return 0, err
	}

	if collection == nil {
		return 0, fmt.Errorf("No Scan Result Exists")
	}

	// Perform the update operation
	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
		return 0, fmt.Errorf("No Document Updated")
	}

	return result.ModifiedCount, nil
}

// IPListByAlert retrieves a list of distinct IPs from the collection based on the provided alert title and target ID.
func (s *ScanResultRepository) IPListByAlert(ctx context.Context, alertTitle string, targetID primitive.ObjectID) ([]string, error) {
	filter := bson.M{
		"vulnerability_title": alertTitle,
		"target_id":           targetID,
	}

	// projection := bson.M{"ip": 1}
	// opts := options.FindOne().SetProjection(projection)

	results, err := s.collection.Distinct(ctx, "ip", filter)
	if err != nil {
		return nil, err
	}

	var ips []string

	for _, result := range results {
		ips = append(ips, result.(string))
	}

	return ips, nil
}

// IPListByAlert retrieves a list of distinct IPs from the collection based on the provided alert title and target ID.
func (s *ScanResultRepository) IPListByAlertDifferential(ctx context.Context, alertTitle string, targetID primitive.ObjectID) ([]string, error) {
	filter := bson.M{
		"vulnerability_title": alertTitle,
	}

	// Perform the delete operation
	collection, err := s.getScanResultCollectionByTarget(ctx, targetID)

	if err != nil {
		return nil, err
	}

	if collection == nil {
		return nil, fmt.Errorf("No Scan Result Exists")
	}

	results, err := collection.Distinct(ctx, "ip", filter)
	if err != nil {
		return nil, err
	}

	var ips []string

	for _, result := range results {
		ips = append(ips, result.(string))
	}

	return ips, nil
}

// Delete Scan result collection by target_id
func (s *ScanResultRepository) DeleteScanResultByTargetID(ctx context.Context, target_id primitive.ObjectID) (int, error) {
	collection := s.getCollectionByTarget(target_id)

	// Drop the collection
	if err := collection.Drop(ctx); err != nil {
		return 0, fmt.Errorf("Error Deleting Scan Results")
	}
	return 1, nil
}

// Delete Scan result entry by _id, we also need target_id to get the collection we search scan result in (scan_results_{target_id})
func (s *ScanResultRepository) DeleteScanByID(ctx context.Context, target_id primitive.ObjectID, scan_number int) (int, error) {
	// Construct the filter for documents to delete
	filter := bson.D{
		{Key: "$and", Value: bson.A{
			bson.D{
				{Key: "scan_numbers", Value: bson.D{
					{Key: "$not", Value: bson.D{
						{Key: "$in", Value: bson.A{scan_number}},
					}},
				}},
			},
			bson.D{
				{Key: "alert_status", Value: bson.D{
					{Key: "$nin", Value: bson.A{enums.AlertFalsePositive, enums.AlertIgnored}},
				}},
			},
		}},
	}

	// Perform the delete operation
	collection, err := s.getScanResultCollectionByTarget(ctx, target_id)
	if err != nil {
		return 0, err
	}

	if collection == nil {
		return 0, fmt.Errorf("No Scan Result Exists")
	}
	result, err := collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
		return 0, fmt.Errorf("Error Deleting Scan")
	}

	return int(result.DeletedCount), nil
}

func (s *ScanResultRepository) GetDetailAlertsByTarget(ctx context.Context, target_id string) ([]datamodels.ScanResult, error) {
	var scanResults []datamodels.ScanResult

	// Timeout context for the DB query
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	targetID, err := primitive.ObjectIDFromHex(target_id)
	if err != nil {
		return nil, errors.New("Invalid ObjectID")
	}

	collection, err := s.getScanResultCollectionByTarget(ctx, targetID)
	if err != nil {
		return nil, err
	}

	if collection == nil {
		return nil, fmt.Errorf("No Scan Result Exists")
	}
	filter := bson.D{
		{Key: "alert_status", Value: bson.D{
			{Key: "$nin", Value: bson.A{enums.AlertFalsePositive, enums.AlertIgnored}},
		}},
	}

	options := options.Find()

	options.SetSort(bson.D{
		{Key: "severity", Value: 1},
	})
	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		return scanResults, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var scanResult datamodels.ScanResult
		err := cursor.Decode(&scanResult)
		if err != nil {
			return scanResults, err
		}
		scanResults = append(scanResults, scanResult)
	}

	return scanResults, nil
}
