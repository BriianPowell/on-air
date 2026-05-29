package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brianpowell/on-air/internal/config"
	"github.com/brianpowell/on-air/internal/detect"
)

type Publisher interface {
	Publish(onAir bool, status detect.Status) error
}

type Agent struct {
	cfg           config.Config
	detector      detect.Detector
	publisher     Publisher
	onAir         bool
	lastPublished detect.Status
	published     bool
}

func New(cfg config.Config, detector detect.Detector, publisher Publisher) *Agent {
	return &Agent{
		cfg:       cfg,
		detector:  detector,
		publisher: publisher,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()

	var (
		activeSince     time.Time
		idleSince       time.Time
		haveActiveSince bool
		haveIdleSince   bool
	)

	if err := a.publish(false, detect.Status{}); err != nil {
		log.Printf("initial publish failed: %v", err)
	} else {
		a.published = true
		a.lastPublished = detect.Status{}
	}

	for {
		status, err := a.detector.Poll()
		if err != nil {
			return fmt.Errorf("poll detector: %w", err)
		}

		nextOnAir, onAirChanged := a.nextState(status, &activeSince, &idleSince, &haveActiveSince, &haveIdleSince)
		if a.shouldPublish(nextOnAir, status, onAirChanged) {
			if err := a.publish(nextOnAir, status); err != nil {
				log.Printf("publish failed: %v", err)
			} else {
				a.onAir = nextOnAir
				a.lastPublished = status
				a.published = true
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (a *Agent) nextState(
	status detect.Status,
	activeSince *time.Time,
	idleSince *time.Time,
	haveActiveSince *bool,
	haveIdleSince *bool,
) (bool, bool) {
	now := time.Now()

	if status.OnAir() {
		*haveIdleSince = false
		if !*haveActiveSince {
			*activeSince = now
			*haveActiveSince = true
		}

		if !a.onAir && now.Sub(*activeSince) >= a.cfg.OnDebounce {
			return true, true
		}
		return a.onAir, false
	}

	*haveActiveSince = false
	if !*haveIdleSince {
		*idleSince = now
		*haveIdleSince = true
	}

	if a.onAir && now.Sub(*idleSince) >= a.cfg.OffDebounce {
		return false, true
	}

	return a.onAir, false
}

func (a *Agent) shouldPublish(nextOnAir bool, status detect.Status, onAirChanged bool) bool {
	if !a.published {
		return true
	}
	if onAirChanged {
		return true
	}
	// Republish mic/camera changes while still on-air internally.
	if nextOnAir && status != a.lastPublished {
		return true
	}
	return false
}

func (a *Agent) publish(onAir bool, status detect.Status) error {
	if a.publisher == nil {
		return nil
	}
	return a.publisher.Publish(onAir, status)
}

// PollOnce returns the current debounced on-air state without waiting.
func (a *Agent) PollOnce() (detect.Status, bool, error) {
	status, err := a.detector.Poll()
	if err != nil {
		return detect.Status{}, false, err
	}

	var activeSince, idleSince time.Time
	haveActiveSince := false
	haveIdleSince := false

	onAir, _ := a.nextState(status, &activeSince, &idleSince, &haveActiveSince, &haveIdleSince)
	return status, onAir, nil
}
