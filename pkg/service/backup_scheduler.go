package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/abhishekdwivedi3060/aerospike-backup-service/pkg/model"
	"github.com/reugn/go-quartz/quartz"
)

const (
	quartzGroupBackupFull        = "full"
	quartzGroupBackupIncremental = "incremental"
)

var jobStore = &backupJobs{jobs: make(map[string]*quartz.JobDetail)}

type backupJobs struct {
	sync.Mutex
	jobs map[string]*quartz.JobDetail
}

func (b *backupJobs) put(key string, value *quartz.JobDetail) {
	b.Lock()
	defer b.Unlock()
	b.jobs[key] = value
}

// NewAdHocFullBackupJobForRoutine returns a new full backup job for the routine name.
func NewAdHocFullBackupJobForRoutine(name string) *quartz.JobDetail {
	jobStore.Lock()
	defer jobStore.Unlock()
	key := quartz.NewJobKeyWithGroup(name, quartzGroupBackupFull).String()
	job := jobStore.jobs[key]
	if job == nil {
		return nil
	}
	jobKey := quartz.NewJobKeyWithGroup(fmt.Sprintf("%s-adhoc-%d", name, time.Now().UnixMilli()),
		quartzGroupBackupFull)
	return quartz.NewJobDetail(job.Job(), jobKey)
}

func ApplyNewConfig(scheduler quartz.Scheduler, config *model.Config, backends BackendsHolder) error {
	err := scheduler.Clear()
	if err != nil {
		return err
	}

	backends.SetData(BuildBackupBackends(config))

	return scheduleRoutines(scheduler, config, backends)
}

// ScheduleBackup creates a new quartz.Scheduler, schedules all the configured backup jobs,
// starts and returns the scheduler.
func ScheduleBackup(ctx context.Context, config *model.Config, backends BackendsHolder,
) (quartz.Scheduler, error) {
	scheduler := quartz.NewStdScheduler()
	scheduler.Start(ctx)

	err := scheduleRoutines(scheduler, config, backends)
	if err != nil {
		return nil, err
	}
	return scheduler, nil
}

func scheduleRoutines(scheduler quartz.Scheduler, config *model.Config, backends BackendsHolder) error {
	for routineName, routine := range config.BackupRoutines {
		backend, _ := backends.Get(routineName)
		handler, err := newBackupHandler(config, routineName, backend)
		if err != nil {
			slog.Error("failed to create backup handler", "routine", routineName, "err", err)
			continue
		}

		// schedule full backup job for the routine
		if err := scheduleFullBackup(scheduler, handler, routine, routineName); err != nil {
			return err
		}

		if routine.IncrIntervalCron != "" {
			// schedule incremental backup job for the routine
			if err := scheduleIncrementalBackup(scheduler, handler, routine, routineName); err != nil {
				return err
			}
		}
	}
	return nil
}

func scheduleFullBackup(scheduler quartz.Scheduler, handler *BackupHandler,
	routine *model.BackupRoutine, routineName string) error {
	fullCronTrigger, err := quartz.NewCronTrigger(routine.IntervalCron)
	if err != nil {
		return err
	}
	fullJob := newBackupJob(handler, quartzGroupBackupFull)
	fullJobDetail := quartz.NewJobDetail(
		fullJob,
		quartz.NewJobKeyWithGroup(routineName, quartzGroupBackupFull),
	)
	if err = scheduler.ScheduleJob(fullJobDetail, fullCronTrigger); err != nil {
		return err
	}
	jobStore.put(fullJobDetail.JobKey().String(), fullJobDetail)
	if needToRunFullBackupNow(handler.state.LastFullRun, fullCronTrigger) {
		slog.Debug("Schedule initial full backup", "name", routineName)
		fullJobDetail := quartz.NewJobDetail(
			fullJob,
			quartz.NewJobKey(routineName),
		)
		if err = scheduler.ScheduleJob(fullJobDetail, quartz.NewRunOnceTrigger(0)); err != nil {
			return err
		}
	}
	return nil
}

func scheduleIncrementalBackup(scheduler quartz.Scheduler, handler *BackupHandler,
	routine *model.BackupRoutine, routineName string) error {
	incrCronTrigger, err := quartz.NewCronTrigger(routine.IncrIntervalCron)
	if err != nil {
		return err
	}
	incrementalJob := newBackupJob(handler, quartzGroupBackupIncremental)
	incrJobDetail := quartz.NewJobDetail(
		incrementalJob,
		quartz.NewJobKeyWithGroup(routineName, quartzGroupBackupIncremental),
	)
	if err = scheduler.ScheduleJob(incrJobDetail, incrCronTrigger); err != nil {
		return err
	}
	jobStore.put(incrJobDetail.JobKey().String(), incrJobDetail)
	return nil
}

func needToRunFullBackupNow(lastFullRun time.Time, trigger *quartz.CronTrigger) bool {
	if lastFullRun.Equal(time.Time{}) {
		return true // no previous run
	}

	fireTimeNano, err := trigger.NextFireTime(lastFullRun.UnixNano())
	if err != nil {
		return true // some error, run backup to be safe
	}
	if time.Unix(0, fireTimeNano).Before(time.Now()) {
		return true // next scheduled backup is in past
	}

	return false
}
