package repositories

import "go.mongodb.org/mongo-driver/mongo"

type Repository struct {
	ScanResult ScanResultRepo
	DSResult   DSResultRepo
	Target     TargetRepository
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		ScanResult: NewScanResultRepo(db),
		DSResult:   NewDSResultRepo(db),
		Target:     NewTargetRepository(db),
	}
}
