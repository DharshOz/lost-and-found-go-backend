package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FoundItem struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LostPerson       primitive.ObjectID `bson:"lostPerson,omitempty" json:"lostPerson"`
	FoundPerson      primitive.ObjectID `bson:"foundPerson,omitempty" json:"foundPerson"` // Same as FoundBy in routes
	FoundPersonPhone string             `bson:"foundPersonPhone" json:"foundPersonPhone"`
	LocationFound    string             `bson:"locationFound" json:"locationFound"`
	DateFound        time.Time          `bson:"dateFound" json:"dateFound"`
	Name             string             `bson:"name" json:"name"`
	Image            string             `bson:"image,omitempty" json:"image"` // Same as ImageURL in routes
	Description      string             `bson:"description" json:"description"`
	Found            bool               `bson:"found" json:"found"`
	CreatedAt        time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt        time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`

	// Add these for response only (not stored in DB)
	FoundByUser    *User `bson:"-" json:"foundByUser,omitempty"`
	LostPersonUser *User `bson:"-" json:"lostPersonUser,omitempty"`
}

// In FoundItem model file
func (f *FoundItem) ImageURL() string {
	return f.Image
}

func (f *FoundItem) FoundBy() primitive.ObjectID {
	return f.FoundPerson
}
