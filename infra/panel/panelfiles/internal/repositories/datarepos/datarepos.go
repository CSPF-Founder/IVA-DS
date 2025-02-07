package datarepos

import "go.mongodb.org/mongo-driver/mongo"

type Repository struct {
	DB         *mongo.Database
	Target     TargetRepository
	ScanResult ScanResultRepository
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		DB:         db,
		Target:     NewTargetRepository(db),
		ScanResult: NewScanResultRepository(db),
	}
}
