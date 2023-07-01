package models

type Role struct {
	// role id
	ID uint `gorm:"primarykey"`

	// role name, e.g. "admin or 123456"
	Name string ` gorm:"index"`

	// role status
	Status string

	// org id,
	OrgId string `gorm:"index"`

	// role display name, e.g. "administrators"
	DisplayName string
}
