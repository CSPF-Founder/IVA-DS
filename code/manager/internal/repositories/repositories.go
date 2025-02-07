package repositories

import "go.mongodb.org/mongo-driver/mongo"

type Repository struct {
	Target TargetRepository
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		Target: NewTargetRepository(db),
	}
}
