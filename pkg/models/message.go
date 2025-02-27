package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message represents Telegram exported-to-HTML message
type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UUID      string             `bson:"uuid" json:"uuid"`
	MessageID string             `bson:"message_id" json:"messageID"`
	Group     string             `bson:"group" json:"group"`
	Datetime  time.Time          `bson:"datetime" json:"datetime"`
	Tags      []string           `bson:"tags" json:"tags"`
}
