package db

import (
	"context"

	commonErrors "github.com/kiga-hub/arc/error"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jinzhu/gorm"
)

const (
	// ConfigKeyDBVersion is the key for db version in config table
	ConfigKeyDBVersion = "db_version"
)

// Config -
type Config struct {
	Name        string `json:"name" gorm:"primary_key;type:varchar(32)"`
	Value       string `json:"value" gorm:"type:varchar(128)"`
	Type        string `json:"type" gorm:"type:varchar(100)"`
	Description string `json:"description" gorm:"type:text"`
}

// Validate -
func (c *Config) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required, validation.In(ConfigKeyDBVersion)),
	)
}

// Add -
func (c *Config) Add(ctx context.Context, db *gorm.DB) (int64, error) {
	if db == nil {
		return 0, commonErrors.ErrParams
	}
	r := db.Where(Config{Name: c.Name}).FirstOrCreate(c)
	return r.RowsAffected, r.Error
}

// Update -
func (c *Config) Update(ctx context.Context, db *gorm.DB) (int64, error) {
	if db == nil || c.Name == "" {
		return 0, commonErrors.ErrParams
	}
	if err := c.GetByName(ctx, db); err != nil {
		return c.Add(ctx, db)
	}
	r := db.Save(c)
	return r.RowsAffected, r.Error
}

// GetByName -
func (c *Config) GetByName(ctx context.Context, db *gorm.DB) error {
	if db == nil || c.Name == "" {
		return commonErrors.ErrParams
	}
	r := db.Table("config").
		Where("name = ?", c.Name).
		First(c)
	return r.Error
}
