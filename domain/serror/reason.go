package serror

import "net/http"

type Reason struct {
	HTTPStatus uint
	Code       uint
	Message    string
}

var (
	UnknownReason = Reason{
		HTTPStatus: http.StatusInternalServerError,
		Code:       0,
		Message:    "unknown_error_occur",
	}
	FailToDecodeBody = Reason{
		HTTPStatus: http.StatusBadRequest,
		Code:       1,
		Message:    "fail_to_decode",
	}
)

var (
	FailToHash = Reason{
		HTTPStatus: http.StatusInternalServerError,
		Code:       2,
		Message:    "internal_error",
	}
	InvalidSigningKey = Reason{
		HTTPStatus: http.StatusInternalServerError,
		Code:       3,
		Message:    "internal_error",
	}
	InvalidTokenClaims = Reason{
		HTTPStatus: http.StatusInternalServerError,
		Code:       4,
		Message:    "internal_error",
	}
	FailToSignToken = Reason{
		HTTPStatus: http.StatusInternalServerError,
		Code:       5,
		Message:    "internal_error",
	}
)

var (
	InvalidUsername = Reason{
		HTTPStatus: http.StatusUnprocessableEntity,
		Code:       2,
		Message:    "invalid_username",
	}
	InvalidEmail = Reason{
		HTTPStatus: http.StatusUnprocessableEntity,
		Code:       3,
		Message:    "invalid_email",
	}
	InvalidDeviceFingerprint = Reason{
		HTTPStatus: http.StatusUnprocessableEntity,
		Code:       4,
		Message:    "invalid_fingerprint",
	}
	DuplicatedEmail = Reason{
		HTTPStatus: http.StatusConflict,
		Code:       5,
		Message:    "duplicated_email",
	}
	DuplicatedUsername = Reason{
		HTTPStatus: http.StatusConflict,
		Code:       6,
		Message:    "duplicated_username",
	}
	DuplicatedFingerprint = Reason{
		HTTPStatus: http.StatusConflict,
		Code:       7,
		Message:    "duplicated_fingerprint",
	}
	InvalidPasswordHash = Reason{
		HTTPStatus: http.StatusInternalServerError,
		Code:       8,
		Message:    "internal_error",
	}
)

var (
	InvalidCredential = Reason{
		HTTPStatus: http.StatusUnauthorized,
		Code:       2,
		Message:    "invalid_credential",
	}
)
