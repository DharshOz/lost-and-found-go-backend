package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LostItem struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Category    string             `bson:"category" json:"category"`
	ImageURL    string             `bson:"image" json:"image"` // Assuming this is intended as image URL
	District    string             `bson:"district" json:"district"`
	State       string             `bson:"state" json:"state"`

	Locations []string `bson:"locations,omitempty" json:"locations,omitempty"` // ✅ NEW FIELD

	DateLost      time.Time          `bson:"dateLost" json:"dateLost"`
	CreatedBy     primitive.ObjectID `bson:"user" json:"createdBy"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
	CreatedByUser *User              `bson:"-" json:"createdByUser,omitempty"`
}
