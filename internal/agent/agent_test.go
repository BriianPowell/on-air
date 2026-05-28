package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianpowell/on-air/internal/config"
	"github.com/brianpowell/on-air/internal/detect"
)

type fakeDetector struct {
	status detect.Status
}

func (f fakeDetector) Poll() (detect.Status, error) {
	return f.status, nil
}

type recordingPublisher struct {
	events []publishEvent
}

type publishEvent struct {
	onAir  bool
	status detect.Status
}

func (r *recordingPublisher) Publish(onAir bool, status detect.Status) error {
	r.events = append(r.events, publishEvent{onAir: onAir, status: status})
	return nil
}

func TestDebounceTurnsOnAfterDelay(t *testing.T) {
	cfg := config.Default("test")
	cfg.OnDebounce = 50 * time.Millisecond
	cfg.OffDebounce = 50 * time.Millisecond
	cfg.PollInterval = 10 * time.Millisecond

	det := fakeDetector{status: detect.Status{MicActive: true}}
	pub := &recordingPublisher{}
	a := New(cfg, det, pub)

	var (
		activeSince     time.Time
		idleSince       time.Time
		haveActiveSince bool
		haveIdleSince   bool
	)

	onAir, changed := a.nextState(det.status, &activeSince, &idleSince, &haveActiveSince, &haveIdleSince)
	if onAir || changed {
		t.Fatalf("expected no transition immediately, got onAir=%t changed=%t", onAir, changed)
	}

	time.Sleep(cfg.OnDebounce)

	onAir, changed = a.nextState(det.status, &activeSince, &idleSince, &haveActiveSince, &haveIdleSince)
	if !onAir || !changed {
		t.Fatalf("expected on transition after debounce, got onAir=%t changed=%t", onAir, changed)
	}
}

func TestDebounceTurnsOffAfterIdle(t *testing.T) {
	cfg := config.Default("test")
	cfg.OnDebounce = 10 * time.Millisecond
	cfg.OffDebounce = 50 * time.Millisecond

	a := New(cfg, fakeDetector{}, nil)
	a.onAir = true

	var (
		activeSince     time.Time
		idleSince       time.Time
		haveActiveSince bool
		haveIdleSince   bool
	)

	status := detect.Status{}
	onAir, changed := a.nextState(status, &activeSince, &idleSince, &haveActiveSince, &haveIdleSince)
	if !onAir || changed {
		t.Fatalf("expected to stay on immediately after idle begins, got onAir=%t changed=%t", onAir, changed)
	}

	time.Sleep(cfg.OffDebounce)

	onAir, changed = a.nextState(status, &activeSince, &idleSince, &haveActiveSince, &haveIdleSince)
	if onAir || !changed {
		t.Fatalf("expected off transition after debounce, got onAir=%t changed=%t", onAir, changed)
	}
}

func TestShouldPublishOnActiveStateChangeWhileOnAir(t *testing.T) {
	a := New(config.Default("test"), fakeDetector{}, nil)
	a.onAir = true
	a.published = true
	a.lastPublished = detect.Status{CameraActive: true}

	status := detect.Status{MicActive: true, CameraActive: true}
	if !a.shouldPublish(true, status, false) {
		t.Fatal("expected publish when mic becomes active during on-air state")
	}

	status = detect.Status{CameraActive: true}
	if a.shouldPublish(true, status, false) {
		t.Fatal("expected no publish when active state unchanged")
	}

	if !a.shouldPublish(false, detect.Status{}, true) {
		t.Fatal("expected publish on on-air transition to off")
	}
}

func TestShouldNotPublishDuringOffDebounceWindDown(t *testing.T) {
	a := New(config.Default("test"), fakeDetector{}, nil)
	a.onAir = true
	a.published = true
	a.lastPublished = detect.Status{CameraActive: true}

	status := detect.Status{}
	if a.shouldPublish(true, status, false) {
		t.Fatal("expected no publish when hardware idle during off debounce")
	}
}

func TestShouldNotPublishActiveChangeWhileOff(t *testing.T) {
	a := New(config.Default("test"), fakeDetector{}, nil)
	a.onAir = false
	a.published = true
	a.lastPublished = detect.Status{}

	status := detect.Status{CameraActive: true}
	if a.shouldPublish(false, status, false) {
		t.Fatal("expected no publish for active state change before on debounce completes")
	}
}

func TestOffDebounceResetsWhenActivityReturns(t *testing.T) {
	cfg := config.Default("test")
	cfg.OffDebounce = 50 * time.Millisecond

	a := New(cfg, fakeDetector{}, nil)
	a.onAir = true

	var (
		activeSince     time.Time
		idleSince       time.Time
		haveActiveSince bool
		haveIdleSince   bool
	)

	idle := detect.Status{}
	onAir, changed := a.nextState(idle, &activeSince, &idleSince, &haveActiveSince, &haveIdleSince)
	if !onAir || changed {
		t.Fatalf("expected to remain on after idle begins, got onAir=%t changed=%t", onAir, changed)
	}

	time.Sleep(20 * time.Millisecond)

	active := detect.Status{MicActive: true}
	onAir, changed = a.nextState(active, &activeSince, &idleSince, &haveActiveSince, &haveIdleSince)
	if !onAir || changed {
		t.Fatalf("expected to remain on when activity returns, got onAir=%t changed=%t", onAir, changed)
	}

	time.Sleep(cfg.OffDebounce)

	onAir, changed = a.nextState(active, &activeSince, &idleSince, &haveActiveSince, &haveIdleSince)
	if !onAir || changed {
		t.Fatalf("expected to stay on while activity continues, got onAir=%t changed=%t", onAir, changed)
	}
}

type sequenceDetector struct {
	statuses []detect.Status
	hold     detect.Status
}

func (s *sequenceDetector) Poll() (detect.Status, error) {
	if len(s.statuses) == 0 {
		return s.hold, nil
	}
	status := s.statuses[0]
	s.statuses = s.statuses[1:]
	s.hold = status
	return status, nil
}

func TestRunPublishSequence(t *testing.T) {
	cfg := config.Default("test")
	cfg.PollInterval = 10 * time.Millisecond
	cfg.OnDebounce = 30 * time.Millisecond
	cfg.OffDebounce = 30 * time.Millisecond

	// Need enough camera-only polls that on-debounce fires before mic activates.
	// With 10ms poll interval and 30ms debounce, the 4th poll (~30ms) triggers
	// on_air while still reading camera-only statuses from the sequence.
	cameraOnly := make([]detect.Status, 6)
	for i := range cameraOnly {
		cameraOnly[i] = detect.Status{CameraActive: true}
	}

	det := &sequenceDetector{
		statuses: append(append(cameraOnly,
			detect.Status{CameraActive: true, MicActive: true},
		), detect.Status{}, detect.Status{}, detect.Status{}, detect.Status{}),
	}
	pub := &recordingPublisher{}
	a := New(cfg, det, pub)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := a.Run(ctx); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	if len(pub.events) < 3 {
		t.Fatalf("expected at least 3 publishes, got %d: %+v", len(pub.events), pub.events)
	}

	if pub.events[0].onAir {
		t.Fatalf("initial publish should be off, got %+v", pub.events[0])
	}

	foundOn := false
	foundMidCall := false
	foundOff := false
	for _, event := range pub.events[1:] {
		if event.onAir && event.status.CameraActive && event.status.MicActive {
			foundMidCall = true
		}
		if event.onAir && event.status.CameraActive && !event.status.MicActive {
			foundOn = true
		}
		if !event.onAir {
			foundOff = true
		}
	}

	if !foundOn {
		t.Fatalf("expected debounced on publish with camera only, got: %+v", pub.events)
	}
	if !foundMidCall {
		t.Fatalf("expected mid-call republish when mic became active, got: %+v", pub.events)
	}
	if !foundOff {
		t.Fatalf("expected debounced off publish, got: %+v", pub.events)
	}
}
