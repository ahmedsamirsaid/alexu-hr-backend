package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/usecases"
)

type Scheduler struct {
	autoRejectUC  *usecases.AutoRejectExpiredRequestsUseCase
	holidaySyncUC *usecases.SyncEgyptPublicHolidaysUseCase
	interval      time.Duration
	stopCh        chan struct{}
	doneCh        chan struct{}
}

func New(
	autoRejectUC *usecases.AutoRejectExpiredRequestsUseCase,
	holidaySyncUC *usecases.SyncEgyptPublicHolidaysUseCase,
	intervalHours int,
) *Scheduler {
	return &Scheduler{
		autoRejectUC:  autoRejectUC,
		holidaySyncUC: holidaySyncUC,
		interval:      time.Duration(intervalHours) * time.Hour,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("scheduler.Start", "interval", s.interval)

	// Run immediately on start
	s.runJobs(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	go func() {
		defer close(s.doneCh)
		for {
			select {
			case <-ticker.C:
				s.runJobs(ctx)
			case <-s.stopCh:
				slog.Info("scheduler.Stop.received")
				return
			case <-ctx.Done():
				slog.Info("scheduler.Stop.context_cancelled")
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	slog.Info("scheduler.Stop.initiating")
	close(s.stopCh)
	<-s.doneCh
	slog.Info("scheduler.Stop.complete")
}

func (s *Scheduler) runJobs(ctx context.Context) {
	slog.Debug("scheduler.runJobs.start")

	if s.autoRejectUC != nil {
		output, err := s.autoRejectUC.Execute(ctx)
		if err != nil {
			slog.Error("scheduler.runJobs.auto_reject", "error", err)
		} else if output.RejectedCount > 0 {
			slog.Info("scheduler.runJobs.auto_reject.completed", "rejected_count", output.RejectedCount)
		}
	}

	if s.holidaySyncUC != nil {
		output, err := s.holidaySyncUC.Execute(ctx)
		if err != nil {
			slog.Error("scheduler.runJobs.holiday_sync", "error", err)
		} else if output.CreatedCount > 0 || output.UpdatedCount > 0 || output.DeletedAbsenceCount > 0 {
			slog.Info(
				"scheduler.runJobs.holiday_sync.completed",
				"created", output.CreatedCount,
				"updated", output.UpdatedCount,
				"deleted_absences", output.DeletedAbsenceCount,
				"skipped_past", output.SkippedPastCount,
			)
		}
	}

	slog.Debug("scheduler.runJobs.complete")
}
