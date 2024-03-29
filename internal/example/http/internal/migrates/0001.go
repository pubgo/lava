package migrates

import (
	"github.com/pubgo/lava/core/migrates"
	"gorm.io/gorm"
)

func m0001() *migrates.Migration {
	type Action struct {
		Code string `gorm:"size:64;primaryKey"`
		Type string `gorm:"size:8;not null"`
		Name string `gorm:"size:64"`
	}

	type MenuItem struct {
		ID         uint
		Code       string `gorm:"size:64;index"`
		ParentCode string `gorm:"size:64;index"`
		Platform   string `gorm:"size:8"`
	}

	type Endpoint struct {
		ID         uint
		TargetType string `gorm:"size:8"`
		Method     string `gorm:"size:8"`
		Path       string `gorm:"size:256"`
		ApiCode    string `gorm:"size:64;index"`
		Action     Action `gorm:"foreignkey:ApiCode;references:Code"`
	}

	type Role struct {
		ID          uint   `gorm:"primarykey"`
		Name        string `gorm:"index;size:8"`
		Status      string `gorm:"size:8"`
		OrgId       string `gorm:"index;size:8"`
		DisplayName string `gorm:"size:32"`
	}

	return &migrates.Migration{
		ID: "0001_action",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&Action{}, &MenuItem{}, &Endpoint{}, &Role{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&Action{}, &MenuItem{}, &Endpoint{}, &Role{})
		},
	}
}
