package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/abhishekdwivedi3060/aerospike-backup-service/pkg/util"
	"github.com/reugn/go-quartz/quartz"
)

// backupJob implements the quartz.Job interface.
type backupJob struct {
	handler   *BackupHandler
	jobType   string
	isRunning atomic.Bool
}

var _ quartz.Job = (*backupJob)(nil)

// Execute is called by a Scheduler when the Trigger associated with this job fires.
func (j *backupJob) Execute(ctx context.Context) error {
	if j.isRunning.CompareAndSwap(false, true) {
		defer j.isRunning.Store(false)
		switch j.jobType {
		case quartzGroupBackupFull:
			j.handler.runFullBackup(time.Now())
		case quartzGroupBackupIncremental:
			j.handler.runIncrementalBackup(time.Now())
		default:
			slog.Error("Unsupported backup type",
				"type", j.jobType,
				"name", j.handler.routineName)
		}
	} else {
		slog.Log(ctx, util.LevelTrace,
			"Backup is currently in progress, skipping it",
			"type", j.jobType,
			"name", j.handler.routineName)
		incrementSkippedCounters(j.jobType)
	}
	return nil
}

func incrementSkippedCounters(jobType string) {
	switch jobType {
	case quartzGroupBackupFull:
		backupSkippedCounter.Inc()
	case quartzGroupBackupIncremental:
		incrBackupSkippedCounter.Inc()
	}
}

// Description returns the description of the backup job.
func (j *backupJob) Description() string {
	return fmt.Sprintf("%s %s backup job", j.handler.routineName, j.jobType)
}

// newBackupJob creates a new backup job.
func newBackupJob(handler *BackupHandler, jobType string) quartz.Job {
	return &backupJob{
		handler: handler,
		jobType: jobType,
	}
}
