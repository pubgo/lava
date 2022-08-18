package result

type Data[T any] struct {
	Body      T      `json:"body,omitempty"`
	ErrMsg    string `json:"err_msg,omitempty"`
	ErrDetail string `json:"err_detail,omitempty"`
}
