package datarepos

import (
	"context"
	"fmt"
	"log"

	"github.com/CSPF-Founder/iva/panel/enums"
	"github.com/CSPF-Founder/iva/panel/models/datamodels"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DSResultRepo struct {
	collection *mongo.Collection
}

func NewDSResultRepo(db *mongo.Database, targetID primitive.ObjectID) DSResultRepo {
	return DSResultRepo{
		collection: db.Collection(fmt.Sprintf("scan_results_%s", targetID.Hex())),
	}
}

// Get Scan Results data by target_id, isDs to check if we have to get data from scan_results collection
// or scan_results_{target_id} collection
func (s *DSResultRepo) ListByTarget(ctx context.Context, target_id primitive.ObjectID, isDs bool) ([]datamodels.ScanResult, error) {
	var scanResults []datamodels.ScanResult

	options := options.Find()
	options.SetSort(bson.D{
		{Key: "severity", Value: 1},
		{Key: "alert_status", Value: 1},
	})

	filter := map[string]any{}

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

func (s *DSResultRepo) ListByAlertStatus(
	ctx context.Context,
	alertStatus enums.AlertStatus,
) ([]datamodels.ScanResult, error) {

	var scanResults []datamodels.ScanResult

	options := options.Find()
	options.SetSort(bson.D{
		{Key: "severity", Value: 1},
	})

	filter := map[string]any{
		"alert_status": alertStatus,
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
// func (s *DSResultRepo) dbHasCollection(ctx context.Context, collectionName string) (bool, error) {
// 	coll, err := s.db.ListCollectionNames(ctx, bson.D{{Key: "name", Value: collectionName}})
// 	if err != nil {
// 		return false, fmt.Errorf("Error Getting Scan Results!")
// 	}
// 	return len(coll) == 1, nil
// }

func (r *DSResultRepo) ByID(ctx context.Context, id primitive.ObjectID) (datamodels.ScanResult, error) {
	var scanResult datamodels.ScanResult
	filter := map[string]any{
		"_id": id,
	}

	cursor := r.collection.FindOne(ctx, filter)
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
func (r *DSResultRepo) UpdateAlertStatus(ctx context.Context, id primitive.ObjectID, targetId primitive.ObjectID, status enums.AlertStatus) (int64, error) {
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

	// Perform the update operation
	result, err := r.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
		return 0, fmt.Errorf("No Document Updated")
	}

	return result.ModifiedCount, nil
}

// IPListByAlert retrieves a list of distinct IPs from the collection based on the provided alert title and target ID.
func (r *DSResultRepo) IPListByAlert(ctx context.Context, alertTitle string, targetID primitive.ObjectID) ([]string, error) {
	filter := bson.M{
		"vulnerability_title": alertTitle,
		"target_id":           targetID,
	}

	// projection := bson.M{"ip": 1}
	// opts := options.FindOne().SetProjection(projection)

	results, err := r.collection.Distinct(ctx, "ip", filter)
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
func (r *DSResultRepo) IPListByAlertDifferential(ctx context.Context, alertTitle string, targetID primitive.ObjectID) ([]string, error) {
	filter := bson.M{
		"vulnerability_title": alertTitle,
	}

	results, err := r.collection.Distinct(ctx, "ip", filter)
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
func (r *DSResultRepo) DeleteScanResult(ctx context.Context) (int, error) {
	// Drop the collection
	if err := r.collection.Drop(ctx); err != nil {
		return 0, fmt.Errorf("Error Deleting Scan Results")
	}
	return 1, nil
}

// Delete Scan result entry by _id, we also need target_id to get the collection we search scan result in (scan_results_{target_id})
func (r *DSResultRepo) DeleteScanByID(ctx context.Context, scan_number int) (int, error) {
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

	result, err := r.collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
		return 0, fmt.Errorf("Error Deleting Scan")
	}

	return int(result.DeletedCount), nil
}

func (r *DSResultRepo) GetDetailAlertsByTarget(ctx context.Context) ([]datamodels.ScanResult, error) {
	var scanResults []datamodels.ScanResult

	filter := bson.D{
		{Key: "alert_status", Value: bson.D{
			{Key: "$nin", Value: bson.A{enums.AlertFalsePositive, enums.AlertIgnored}},
		}},
	}

	options := options.Find()

	options.SetSort(bson.D{
		{Key: "severity", Value: 1},
	})
	cursor, err := r.collection.Find(ctx, filter, options)
	if err != nil {
		defer cursor.Close(ctx)
		return scanResults, err
	}

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
