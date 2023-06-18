package accounts

import "github.com/kamva/mgm/v3"

type Account struct {
	mgm.DefaultModel `bson:",inline"`

	ID         string   `json:"id" bson:"_id"`
	Name       string   `json:"name"`
	Password   string   `json:"password"`
	Email      string   `json:"email"`
	Slug       string   `json:"slug"`
	JoinedOn   string   `json:"joinedOn" bson:"created_at"`
	UpdatedOn  string   `json:"updatedOn" bson:"updated_at"`
	LastActive string   `json:"lastActive" bson:"last_active"`
	Blocked    []string `json:"blocked" bson:"blocked"`
}
