package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/CSPF-Founder/iva/manager/enums"
	"github.com/CSPF-Founder/iva/manager/models"
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

func (c *TargetRepository) FindById(ctx context.Context, targetID primitive.ObjectID) (*models.Target, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if targetID.IsZero() {
		return nil, errors.New("invalid mongodb object id given")
	}

	var target models.Target
	err := c.collection.FindOne(ctx, bson.M{"_id": targetID}).Decode(&target)
	if err != nil {
		return nil, err
	}

	return &target, nil
}

func (c *TargetRepository) UpdateScanStatus(ctx context.Context, target *models.Target) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	now := time.Now()
	toUpdate := bson.M{}

	switch target.ScanStatus {
	case enums.TargetStatusInitiatingScan:
		target.ScanInitiatedTime = &now
		toUpdate["scan_initiated_time"] = now
	case enums.TargetStatusScanStarted:
		target.ScanStartedTime = &now
		toUpdate["scan_started_time"] = now
	case enums.TargetStatusReportGenerated:
		target.ScanCompletedTime = &now
		toUpdate["scan_completed_time"] = now
	}

	toUpdate["scan_status"] = target.ScanStatus
	_, err := c.collection.UpdateOne(
		ctx,
		bson.M{"_id": target.ID},
		bson.M{"$set": toUpdate},
	)
	return err
}

func (c *TargetRepository) GetNetworkScanJob(ctx context.Context) (*models.Target, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	filters := bson.M{
		"target_type": bson.M{
			"$in": []string{"ip", "ip_range"},
		},
		"$or": []bson.M{
			{"scan_status": enums.TargetStatusYetToStart},
			{"scan_status": bson.M{"$exists": false}},
		},
	}

	var target models.Target
	err := c.collection.FindOne(ctx, filters).Decode(&target)
	if err != nil {
		return nil, err
	}

	return &target, nil
}

func (c *TargetRepository) GetWebScanJob(ctx context.Context) (*models.Target, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	filters := bson.M{
		"target_type": bson.M{
			"$in": []string{"url"},
		},
		"$or": []bson.M{
			{"scan_status": enums.TargetStatusYetToStart},
			{"scan_status": bson.M{"$exists": false}},
		},
	}

	var target models.Target
	err := c.collection.FindOne(ctx, filters).Decode(&target)
	if err != nil {
		return nil, err
	}

	return &target, nil
}

func (c *TargetRepository) cursorToTargetModels(ctx context.Context, cursor *mongo.Cursor) ([]models.Target, error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var targets []models.Target

	if err := cursor.All(ctx, &targets); err != nil {
		return nil, err
	}

	return targets, nil
}

func (c *TargetRepository) GetJobListToScan(ctx context.Context) ([]models.Target, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	filters := bson.M{}

	filters["flag"] = enums.ScanFlagWaitingToStart
	filters["$or"] = []bson.M{
		{
			"scan_initiated_time": bson.M{
				"$lt": time.Now().Add(-6 * time.Hour),
			},
		},
		{
			"scan_initiated_time": bson.M{"$exists": false},
		},
	}

	cursor, err := c.collection.Find(ctx, filters)
	if err != nil {
		return nil, err
	}

	data, err := c.cursorToTargetModels(ctx, cursor)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *TargetRepository) MarkUnfinishedAsFailed(ctx context.Context, targetTypes []string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	filters := bson.M{
		"scan_status": bson.M{
			"$nin": []enums.TargetStatus{
				enums.TargetStatusYetToStart,
				enums.TargetStatusReportGenerated,
				enums.TargetStatusScanFailed,
				enums.TargetStatusUnreachable,
			},
		},
		"target_type": bson.M{
			"$in": targetTypes,
		},
	}
	toUpdate := bson.M{"scan_status": enums.TargetStatusScanFailed}
	finalFilter := map[string]any{
		"$and": []any{filters},
	}
	_, err := c.collection.UpdateMany(ctx, finalFilter, bson.M{"$set": toUpdate})

	return err
}
