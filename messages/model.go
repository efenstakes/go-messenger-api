package messages

import (
	"time"

	"github.com/kamva/mgm/v3"
)

type Message struct {
	mgm.DefaultModel `bson:",inline"`

	// ID          string `json:"id" bson:"_id"`
	Text        string    `json:"text"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	SendAt      time.Time `json:"sendAt" bson:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updated_at"`
	DeliveredAt string    `json:"deliveredAt"`
}
