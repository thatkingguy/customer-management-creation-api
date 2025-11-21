package models

// Industry represents the industry table
type Industry struct {
	IndustryID  int    `db:"industryId" json:"industryId"`
	SectorID    int    `db:"sectorId" json:"sectorId"`
	Description string `db:"description" json:"description"`
}
