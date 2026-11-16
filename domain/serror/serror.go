package serror

import "fmt"

type ServiceError struct {
	Code       string `json:"code"`
	HTTPStatus uint   `json:"-"`
	Message    string `json:"message,omitempty"`
}

func NewServiceError(category Category, subCategory SubCategory, reason Reason) ServiceError {
	return ServiceError{
		Code:       fmt.Sprintf("%2d%2d%4d", category, subCategory, reason.Code),
		HTTPStatus: reason.HTTPStatus,
		Message:    reason.Message,
	}
}

func (err ServiceError) Error() string {
	return fmt.Sprintf("[%s] (%d) %s", err.Code, err.HTTPStatus, err.Message)
}

func (e ServiceError) Is(target error) bool {
	switch t := target.(type) {
	case ServiceError:
		return e.Code == t.Code
	case *ServiceError:
		return e.Code == t.Code
	default:
		return false
	}
}
