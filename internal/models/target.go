package models

// Target represents the target table
type Target struct {
	TargetID    int     `db:"targetId" json:"targetId"`
	ShortName   string  `db:"shortName" json:"shortName"`     // NOT NULL
	Description *string `db:"description" json:"description"` // nullable
}
