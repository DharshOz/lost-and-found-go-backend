package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Bookmark struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User      primitive.ObjectID `bson:"user,omitempty" json:"user"`
	LostItem  primitive.ObjectID `bson:"lostItem,omitempty" json:"lostItem"`
	CreatedAt time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}
