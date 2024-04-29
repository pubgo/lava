package lava

// Annotation is used to attach arbitrary metadata to the schema objects
type Annotation interface {
	Name() string
}

type Annotations = []Annotation
