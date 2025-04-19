package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Location struct {
	District string `bson:"district" json:"district"`
	State    string `bson:"state" json:"state"`
}

type Notification struct {
	Message   string    `bson:"message" json:"message"`
	Read      bool      `bson:"read" json:"read"`
	CreatedAt time.Time `bson:"createdAt,omitempty" json:"createdAt"`
}

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username      string             `bson:"username" json:"username"`
	Email         string             `bson:"email" json:"email"`
	Phone         string             `bson:"phone" json:"phone"`
	Profession    string             `bson:"profession" json:"profession"`
	Location      Location           `bson:"location" json:"location"`
	Password      string             `bson:"password" json:"password"`
	Notifications []Notification     `bson:"notifications,omitempty" json:"notifications"`
	CreatedAt     time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}
