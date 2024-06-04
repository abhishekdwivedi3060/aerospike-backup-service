package service

import (
	"github.com/abhishekdwivedi3060/aerospike-backup-service/pkg/model"
)

type RestoreService interface {
	// Restore starts a restore process using the given request.
	// Returns the job id as a unique identifier.
	Restore(request *model.RestoreRequestInternal) (int, error)

	// RestoreByTime starts a restore by time process using the given request.
	// Returns the job id as a unique identifier.
	RestoreByTime(request *model.RestoreTimestampRequest) (int, error)

	// JobStatus returns status for the given job id.
	JobStatus(jobID int) (*model.RestoreJobStatus, error)

	// RetrieveConfiguration return backed up Aerospike configuration.
	RetrieveConfiguration(routine string, toTimeMillis int64) ([]byte, error)
}
