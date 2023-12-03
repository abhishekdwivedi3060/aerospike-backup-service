package service

import (
	"testing"

	"github.com/aerospike/backup/pkg/model"
	"github.com/aerospike/backup/pkg/util"
)

func TestRestoreMemory(t *testing.T) {
	restoreService := NewRestoreMemory(nil)
	restoreRequest := &model.RestoreRequest{
		DestinationCuster: &model.AerospikeCluster{
			Host: util.Ptr("localhost"),
			Port: util.Ptr(int32(3000)),
		},
		Policy: &model.RestorePolicy{
			SetList: []string{"set1"},
		},
		Directory: util.Ptr("./testout/backup"),
	}
	jobID := restoreService.Restore(restoreRequest)

	jobStatus := restoreService.JobStatus(jobID)
	if jobStatus != jobStatusRunning {
		t.Errorf("Expected jobStatus to be %s", jobStatusRunning)
	}

	wrongJobStatus := restoreService.JobStatus(1111)
	if wrongJobStatus != jobStatusNA {
		t.Errorf("Expected jobStatus to be %s", jobStatusNA)
	}
}