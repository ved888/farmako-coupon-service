package utils

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
)

var generator *shortid.Shortid
var BuildNumber string

const generatorSeed = 1000

type Branch string

const (
	Production  Branch = "main"
	Staging     Branch = "stage"
	Development Branch = "dev"
)

type Status string

const (
	Activate   Status = "activate"
	Deactivate Status = "deactivate"
	All        Status = "all"
	Break      Status = "break"
)

type AllStatus string

const (
	True  AllStatus = "true"
	False AllStatus = "false"
)

type GenericResponse struct {
	Message string `json:"message"`
} // @name GenericResponse

func Response(w http.ResponseWriter, message string) {
	RespondJSON(w, http.StatusOK, GenericResponse{Message: message})
}

const RegularExpression = "^(?i)(SC_[a-zA-Z0-9_\\-\\.]*|C_[a-zA-Z0-9_\\-\\.]*|T_[a-zA-Z0-9_\\-\\.]*)"

type FieldError struct {
	Err validator.ValidationErrors
}

// RequestErr models contains the body having details related with some kind of error
// which happened during processing of a request
type RequestErr struct {
	// ID for the request
	// Example: 8YeCqPXmM
	ID string `json:"id"`

	// MessageToUser will contain error message
	// Example: Invalid Email
	MessageToUser string `json:"messageToUser"`

	// DeveloperInfo will contain additional developer info related with error
	// Example: Invalid email format
	DeveloperInfo string `json:"developerInfo"`

	// Err contains the error or exception message
	// Example: validation on email failed with error invalid email format
	Err string `json:"error"`

	// StatusCode will contain the status code for the error
	// Example: 500
	StatusCode int `json:"statusCode"`

	// IsClientError will be false if some internal server error occurred
	IsClientError bool `json:"isClientError"`
} // @name RequestErr

func init() {
	n, err := rand.Int(rand.Reader, big.NewInt(generatorSeed))
	if err != nil {
		logrus.Panicf("failed to initialize utilities with random seed, %+v", err)
		return
	}

	g, err := shortid.New(1, shortid.DefaultABC, n.Uint64())

	if err != nil {
		logrus.Panicf("Failed to initialize utils package with error: %+v", err)
	}
	generator = g
}

// ParseBody parses the values from io reader to a given interface
func ParseBody(body io.Reader, out interface{}) error {
	err := json.NewDecoder(body).Decode(out)
	if err != nil {
		return err
	}

	return nil
}

// EncodeJSONBody writes the JSON body to response writer
func EncodeJSONBody(resp http.ResponseWriter, data interface{}) error {
	return json.NewEncoder(resp).Encode(data)
}

// RespondJSON sends the interface as a JSON
func RespondJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if body != nil {
		if err := EncodeJSONBody(w, body); err != nil {
			logrus.Errorf("Failed to respond JSON with error: %+v", err)
		}
	}
}

// newClientError creates structured client error response message
func newClientError(err error, statusCode int, messageToUser string, additionalInfoForDevs ...string) *RequestErr {
	additionalInfoJoined := strings.Join(additionalInfoForDevs, "\n")
	if additionalInfoJoined == "" {
		additionalInfoJoined = messageToUser
	}

	errorID, _ := generator.Generate()
	var errString string
	if err != nil {
		errString = err.Error()
	}
	clientErr := true
	if statusCode == http.StatusInternalServerError {
		clientErr = false
	}
	return &RequestErr{
		ID:            errorID,
		MessageToUser: messageToUser,
		DeveloperInfo: additionalInfoJoined,
		Err:           errString,
		StatusCode:    statusCode,
		IsClientError: clientErr,
	}
}

// RespondError sends an error message to the API caller and logs the error
func RespondError(w http.ResponseWriter, statusCode int, err error, messageToUser string, additionalInfoForDevs ...string) {
	logrus.Errorf("status: %d, message: %s, err: %+v ", statusCode, messageToUser, err)
	clientError := newClientError(err, statusCode, messageToUser, additionalInfoForDevs...)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(clientError); err != nil {
		logrus.Errorf("Failed to send error to caller with error: %+v", err)
	}
}

// IsProd returns true if running on prod
func IsProd() bool {
	return GetBranch() == Production
}

// GetBranch returns current branch name, defaults to development if no branch specified
func GetBranch() Branch {
	b := os.Getenv("BRANCH")
	if b == "" {
		return Development
	}
	return Branch(b)
}

// IsBranchEnvSet checks if the branch environment is set
func IsBranchEnvSet() bool {
	b := os.Getenv("BRANCH")
	return b != ""
}

func GetBuildNumber() string {
	return BuildNumber
}
