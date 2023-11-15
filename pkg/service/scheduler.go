package service

import (
	"context"
	"github.com/aerospike/backup/internal/util"
	"github.com/aerospike/backup/pkg/model"
	"github.com/aerospike/backup/pkg/shared"
	"github.com/aerospike/backup/pkg/stdio"
)

// backup service
var backupService shared.Backup = shared.NewBackup()

// stdIO captures standard output
var stdIO *stdio.CgoStdio = &stdio.CgoStdio{}

// ScheduleHandlers schedules the configured backup policies.
func ScheduleHandlers(ctx context.Context, handlers []BackupScheduler) {
	for _, handler := range handlers {
		handler.Schedule(ctx)
	}
}

// BuildBackupHandlers builds a list of BackupSchedulers according to
// the given configuration.
func BuildBackupHandlers(config *model.Config) []BackupScheduler {
	schedulers := make([]BackupScheduler, 0, len(config.BackupPolicies))
	for _, backupPolicy := range config.BackupPolicies {
		handler, err := NewBackupHandler(config, backupPolicy)
		util.Check(err)
		schedulers = append(schedulers, handler)
	}
	return schedulers
}

// ToBackend returns a list of underlying BackupBackends
// for the given list of BackupSchedulers.
func ToBackend(handlers []BackupScheduler) []BackupBackend {
	backends := make([]BackupBackend, 0, len(handlers))
	for _, scheduler := range handlers {
		backends = append(backends, scheduler.GetBackend())
	}
	return backends
}
