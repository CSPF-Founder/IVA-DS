package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/CSPF-Founder/iva/scanner/enums"
	"github.com/CSPF-Founder/iva/scanner/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TargetRepository struct {
	collection *mongo.Collection
}

func NewTargetRepository(db *mongo.Database) TargetRepository {
	return TargetRepository{
		collection: db.Collection("targets"),
	}
}

func (c *TargetRepository) FindByID(ctx context.Context, targetID string) (target models.Target, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return models.Target{}, err
	}

	err = c.collection.FindOne(ctx, map[string]primitive.ObjectID{"_id": objectID}).Decode(&target)
	// Document found, return the target
	return target, err
}

func (c *TargetRepository) UpdateScanStatus(ctx context.Context, target *models.Target) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if target == nil {
		return false, errors.New("invalid target object given")
	}

	now := time.Now()
	toUpdate := bson.M{}

	switch target.ScanStatus {
	case enums.TargetStatusScanStarted:
		target.ScanStartedTime = now
		toUpdate["scan_started_time"] = now
	case enums.TargetStatusReportGenerated:
		target.ScanCompletedTime = now
		toUpdate["scan_completed_time"] = now
	}

	toUpdate["scan_status"] = target.ScanStatus

	filter := bson.M{"_id": target.ID}
	update := bson.M{"$set": toUpdate}
	_, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *TargetRepository) UpdateScanStatusByID(ctx context.Context, targetID primitive.ObjectID, scanStatus enums.TargetStatus) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	now := time.Now()
	toUpdate := bson.M{}
	switch scanStatus {
	case enums.TargetStatusScanStarted:
		toUpdate["scan_started_time"] = now
	case enums.TargetStatusReportGenerated:
		toUpdate["scan_completed_time"] = now
	}

	filter := bson.M{"_id": targetID}
	update := bson.M{"$set": toUpdate}

	toUpdate["scan_status"] = scanStatus
	_, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *TargetRepository) MarkAsComplete(ctx context.Context, target *models.Target) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	now := time.Now()
	toUpdate := bson.M{}
	toUpdate["scan_completed_time"] = now
	toUpdate["scan_status"] = enums.TargetStatusReportGenerated
	toUpdate["overall_cvss_score"] = target.OverallCVSSScore

	if target.IsIPRange() {
		toUpdate["cvss_score_by_host"] = target.CVSSScoreByHost
	}
	filter := bson.M{"_id": target.ID}
	update := bson.M{"$set": toUpdate}

	_, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *TargetRepository) GetAll(ctx context.Context) (targets []models.Target, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := c.collection.Find(ctx, map[string]string{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &targets); err != nil {
		return nil, err
	}

	return targets, nil
}

// Updates scan numbers For DS
func (c *TargetRepository) UpdateScanInfo(
	ctx context.Context,
	targetID primitive.ObjectID,
	scanInfo models.ScanInfo,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": targetID}
	update := bson.M{
		"$push": bson.M{
			"scans": scanInfo,
		},
	}
	_, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return true, nil
}
