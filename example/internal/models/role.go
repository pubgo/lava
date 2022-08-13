package models

import "github.com/pubgo/lava/example/pkg/proto/permpb"

func RoleFromProto(role *permpb.Role) *Role {
	return &Role{
		ID:          uint(role.Id),
		Name:        role.Name,
		Status:      role.Status,
		OrgId:       role.OrgId,
		DisplayName: role.DisplayName,
	}
}

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

func (t *Role) Proto() *permpb.Role {
	return &permpb.Role{
		Id:          int32(t.ID),
		Name:        t.Name,
		Status:      t.Status,
		OrgId:       t.OrgId,
		DisplayName: t.DisplayName,
	}
}
