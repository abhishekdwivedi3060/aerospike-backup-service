package model

import (
	"encoding/json"
	"errors"
	lib "github.com/aerospike/aerospike-management-lib"
)

// RestoreRequest represents a restore operation request.
// @Description RestoreRequest represents a restore operation request.
type RestoreRequest struct {
	DestinationCuster *AerospikeCluster `json:"destination,omitempty" validate:"required"`
	Policy            *RestorePolicy    `json:"policy,omitempty" validate:"required"`
	SourceStorage     *Storage          `json:"source,omitempty" validate:"required"`
	SecretAgent       *SecretAgent      `json:"secret-agent,omitempty"`
}

// RestoreRequestInternal is used internally to prepopulate data for the restore operation.
type RestoreRequestInternal struct {
	RestoreRequest
	Dir *string
}

// RestoreTimestampRequest represents a restore by timestamp operation request.
// @Description RestoreTimestampRequest represents a restore by timestamp operation request.
type RestoreTimestampRequest struct {
	// The details of the Aerospike destination cluster.
	DestinationCuster *AerospikeCluster `json:"destination,omitempty" validate:"required"`
	// Restore policy to use in the operation.
	Policy *RestorePolicy `json:"policy,omitempty" validate:"required"`
	// Secret Agent configuration (optional).
	SecretAgent *SecretAgent `json:"secret-agent,omitempty"`
	// Required epoch time for recovery. The closest backup before the timestamp will be applied.
	Time int64 `json:"time,omitempty" format:"int64" example:"1739538000000" validate:"required"`
	// The backup routine name.
	Routine string `json:"routine,omitempty" example:"daily" validate:"required"`
}

// String satisfies the fmt.Stringer interface.
func (r RestoreRequest) String() string {
	request, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}
	return string(request)
}

// String satisfies the fmt.Stringer interface.
func (r RestoreTimestampRequest) String() string {
	request, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}
	return string(request)
}

// NewRestoreRequest creates a new RestoreRequest.
func NewRestoreRequest(
	destinationCluster *AerospikeCluster,
	policy *RestorePolicy,
	sourceStorage *Storage,
	secretAgent *SecretAgent,
) *RestoreRequest {
	return &RestoreRequest{
		DestinationCuster: destinationCluster,
		Policy:            policy,
		SourceStorage:     sourceStorage,
		SecretAgent:       secretAgent,
	}
}

// Validate validates the restore operation request.
func (r *RestoreRequest) Validate() error {
	if err := r.DestinationCuster.Validate(); err != nil {
		return err
	}
	if err := r.Policy.Validate(); err != nil {
		return err
	}
	if err := r.SourceStorage.Validate(); err != nil {
		return err
	}
	if err := r.Policy.Validate(); err != nil { //nolint:revive
		return err
	}
	return nil
}

// Validate validates the restore operation request.
func (r *RestoreTimestampRequest) Validate() error {
	if err := r.DestinationCuster.Validate(); err != nil {
		return err
	}
	if err := r.Policy.Validate(); err != nil {
		return err
	}
	if r.Time <= 0 {
		return errors.New("restore point in time should be positive")
	}
	if r.Routine == "" {
		return emptyFieldValidationError(r.Routine)
	}
	return nil
}

func (in *RestoreRequest) DeepCopy() *RestoreRequest {
	if in == nil {
		return nil
	}
	out := new(RestoreRequest)
	in.DeepCopyInto(out)

	return out
}

func (in *RestoreRequest) DeepCopyInto(out *RestoreRequest) {
	temp := lib.DeepCopy(in).(*RestoreRequest)
	*out = *temp
}
