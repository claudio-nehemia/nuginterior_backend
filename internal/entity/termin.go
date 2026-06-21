package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Termin represents the termins table.
type Termin struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	KodeTipe  string    `gorm:"size:255;not null" json:"kode_tipe"`
	NamaTipe  string    `gorm:"size:255;not null" json:"nama_tipe"`
	Deskripsi string    `gorm:"type:text" json:"deskripsi"`
	Tahapan   Tahapans  `gorm:"type:jsonb;not null" json:"tahapan"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Termin) TableName() string {
	return "termins"
}

// TahapanItem represents a single step in the termin tahapan JSONB array.
type TahapanItem struct {
	Step       int     `json:"step"`
	Text       string  `json:"text"`
	Persentase float64 `json:"persentase"`
}

// Tahapans is a custom type for JSONB array of TahapanItem.
type Tahapans []TahapanItem

// Scan implements the sql.Scanner interface for reading from DB.
func (t *Tahapans) Scan(value interface{}) error {
	if value == nil {
		*t = Tahapans{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Tahapans: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, t)
}

// Value implements the driver.Valuer interface for writing to DB.
func (t Tahapans) Value() (driver.Value, error) {
	if t == nil {
		return "[]", nil
	}
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}
