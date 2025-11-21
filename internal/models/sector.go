package models

// Sector represents the sector table
type Sector struct {
	SectorID    int    `db:"sectorId" json:"sectorId"`
	Description string `db:"description" json:"description"`
}
