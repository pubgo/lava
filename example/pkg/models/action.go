package models

import "fmt"

const (
	ActionTypeAPI        = "api"
	ActionTypeMenu       = "menu"
	ActionTypePageAction = "action"
)

// Action represents an operation that can be performed on UI layer or API layer.
type Action struct {
	Code string `gorm:"size:32;primaryKey"`
	Type string `gorm:"size:8;not null"`
	Name string `gorm:"size:64"`
}

func (a *Action) ObjectType() string {
	return ObjectTypeAction
}

func (a *Action) ObjectID() string {
	return fmt.Sprintf("%s/%s", a.Type, a.Code)
}

func (a *Action) String() string {
	return ObjRepr(a)
}

// MenuItem represents a navigation menu or a page action on UI, organized in a tree style.
type MenuItem struct {
	ID         uint
	Code       string `gorm:"size:32;index"`
	ParentCode string `gorm:"size:32;index"`
	Platform   string `gorm:"size:8"`
}

// Endpoint represents a real API via HTTP, WS or some other protocols. It will be mapped to an API action.
type Endpoint struct {
	ID         uint
	TargetType string `gorm:"size:8"`
	Method     string `gorm:"size:8"`
	Path       string `gorm:"size:256"`
	ApiCode    string `gorm:"size:32;index"`
	Action     Action `gorm:"foreignkey:code;references:api_code"`
}
