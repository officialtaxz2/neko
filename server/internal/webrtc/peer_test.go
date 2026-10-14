package webrtc

import (
	"testing"
	"time"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/utils"
)

func adaptiveEstimatorTestConfig() config.WebRTCEstimator {
	return config.WebRTCEstimator{
		StableDuration:             12 * time.Second,
		UnstableDuration:           6 * time.Second,
		StalledDuration:            8 * time.Second,
		DowngradeBackoff:           30 * time.Second,
		UpgradeBackoff:             5 * time.Second,
		RecoveryProbeInterval:      30 * time.Second,
		RecoveryProbeMaxBackoff:    2 * time.Minute,
		DowngradeDeficitThreshold: 0.15,
		TransportReserve:           0.05,
		UpgradeDiffThreshold:       0.15,
	}
}

func TestInitialEstimatorObservationTimesDoNotStartExpired(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	stableSince, unstableSince, stalledSince := initialEstimatorObservationTimes(now)

	for _, tt := range []struct {
		name string
		got  time.Time
	}{
		{name: "stable", got: stableSince},
		{name: "unstable", got: unstableSince},
		{name: "stalled", got: stalledSince},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got.IsZero() {
				t.Fatal("observation window started at zero time")
			}
			if !tt.got.Equal(now) {
				t.Fatalf("observation window = %v, want %v", tt.got, now)
			}
		})
	}
}

func TestNeutralLossFreeEstimateDoesNotDowngradeWithoutUpgradeReserve(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start, conf)
	reference := deliveryBitrateReference(1_900_000, 1_996_800, 128_000, conf.TransportReserve)

	// 2.0 Mbit/s does not provide 15% upgrade-style headroom over the high
	// video tier, but it is above the tolerated-deficit floor for the current
	// complete delivery. A neutral application-limited estimate must hold high.
	for elapsed := 2 * time.Second; elapsed <= 60*time.Second; elapsed += 2 * time.Second {
		decision := state.observe(start.Add(elapsed), utils.TrendDirectionNeutral, 2_000_000, reference, false, conf)
		if decision.insufficient {
			t.Fatalf("estimate classified insufficient at %v", elapsed)
		}
		if decision.downgrade {
			t.Fatalf("unexpected downgrade at %v", elapsed)
		}
	}
}

func TestSevereLossFreeEstimatorCollapseDoesNotDowngrade(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start, conf)
	reference := deliveryBitrateReference(1_996_800, 1_996_800, 128_000, conf.TransportReserve)

	// This reproduces the target-server failure class: GCC can collapse well
	// below the delivery floor even though receiver reports and NACK feedback
	// show no packet congestion. Target value and trend alone are advisory.
	for elapsed := 2 * time.Second; elapsed <= 180*time.Second; elapsed += 2 * time.Second {
		direction := utils.TrendDirectionNeutral
		if elapsed <= 20*time.Second {
			direction = utils.TrendDirectionDownward
		}
		decision := state.observe(start.Add(elapsed), direction, 467_178, reference, false, conf)
		if !decision.insufficient {
			t.Fatalf("collapsed estimate not classified insufficient at %v", elapsed)
		}
		if decision.receiverCongestionConfirmed {
			t.Fatalf("loss-free observation acquired congestion evidence at %v", elapsed)
		}
		if decision.downgrade {
			t.Fatalf("loss-free target collapse caused downgrade at %v", elapsed)
		}
	}
}

func TestReceiverCongestionEvidenceRequiresFreshLossOrNack(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	maxAge := 4 * time.Second

	tests := []struct {
		name     string
		feedback receiverFeedbackSnapshot
		want     bool
	}{
		{
			name: "fresh clean report",
			feedback: receiverFeedbackSnapshot{
				reportAvailable: true,
				reportedAt:      now.Add(-time.Second),
			},
		},
		{
			name: "fresh lossy report",
			feedback: receiverFeedbackSnapshot{
				reportAvailable: true,
				reportedAt:      now.Add(-time.Second),
				fractionLost:    3,
			},
			want: true,
		},
		{
			name: "stale lossy report",
			feedback: receiverFeedbackSnapshot{
				reportAvailable: true,
				reportedAt:      now.Add(-5 * time.Second),
				fractionLost:    3,
			},
		},
		{
			name: "fresh nack",
			feedback: receiverFeedbackSnapshot{
				lastNackAt: now.Add(-time.Second),
			},
			want: true,
		},
		{
			name: "future feedback is rejected",
			feedback: receiverFeedbackSnapshot{
				reportAvailable: true,
				reportedAt:      now.Add(time.Second),
				fractionLost:    3,
				lastNackAt:      now.Add(time.Second),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assessReceiverCongestion(now, tt.feedback, maxAge).confirmed; got != tt.want {
				t.Fatalf("confirmed = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceiverCongestionEvidenceExpiresBeforeUnstableWindow(t *testing.T) {
	if got, want := receiverCongestionEvidenceMaxAge(2*time.Second, 6*time.Second), 4*time.Second; got != want {
		t.Fatalf("tracked feedback age = %v, want %v", got, want)
	}
	if got, want := receiverCongestionEvidenceMaxAge(5*time.Second, 6*time.Second), 6*time.Second-time.Nanosecond; got != want {
		t.Fatalf("clamped feedback age = %v, want %v", got, want)
	}
}

func TestTransientReceiverCongestionDoesNotDowngrade(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start, conf)
	reference := deliveryBitrateReference(1_996_800, 1_996_800, 128_000, conf.TransportReserve)

	for elapsed := 2 * time.Second; elapsed <= 20*time.Second; elapsed += 2 * time.Second {
		confirmed := elapsed == 2*time.Second || elapsed == 4*time.Second ||
			elapsed == 10*time.Second || elapsed == 12*time.Second
		decision := state.observe(
			start.Add(elapsed),
			utils.TrendDirectionDownward,
			1_300_000,
			reference,
			confirmed,
			conf,
		)
		if decision.downgrade {
			t.Fatalf("transient congestion evidence caused downgrade at %v", elapsed)
		}
	}
}

func TestEstimatorCongestionEvidenceRemainsPeerLocal(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	healthy := newEstimatorObservationState(start, conf)
	constrained := newEstimatorObservationState(start, conf)
	reference := deliveryBitrateReference(1_996_800, 1_996_800, 128_000, conf.TransportReserve)

	var constrainedDowngraded bool
	for elapsed := 2 * time.Second; elapsed <= 14*time.Second; elapsed += 2 * time.Second {
		healthyDecision := healthy.observe(
			start.Add(elapsed),
			utils.TrendDirectionNeutral,
			1_300_000,
			reference,
			false,
			conf,
		)
		if healthyDecision.downgrade {
			t.Fatalf("constrained peer affected healthy peer at %v", elapsed)
		}

		constrainedDecision := constrained.observe(
			start.Add(elapsed),
			utils.TrendDirectionNeutral,
			1_300_000,
			reference,
			true,
			conf,
		)
		constrainedDowngraded = constrainedDowngraded || constrainedDecision.downgrade
	}

	if !constrainedDowngraded {
		t.Fatal("sustained peer-local congestion did not downgrade constrained peer")
	}
}

func TestSustainedInsufficientNeutralEstimateDowngrades(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start, conf)
	reference := deliveryBitrateReference(1_996_800, 1_996_800, 128_000, conf.TransportReserve)

	var downgradedAt time.Duration
	for elapsed := 2 * time.Second; elapsed <= 30*time.Second; elapsed += 2 * time.Second {
		decision := state.observe(start.Add(elapsed), utils.TrendDirectionNeutral, 1_300_000, reference, true, conf)
		if decision.downgrade {
			downgradedAt = elapsed
			break
		}
	}

	if got, want := downgradedAt, 14*time.Second; got != want {
		t.Fatalf("first sustained neutral downgrade at %v, want %v", got, want)
	}
}

func TestEstimatorRecoveryRequiresStableCapacityBeforeUpgrade(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start, conf)
	mediumReference := deliveryBitrateReference(748_800, 748_800, 128_000, conf.TransportReserve)
	lowReference := deliveryBitrateReference(332_800, 332_800, 128_000, conf.TransportReserve)
	highReference := deliveryBitrateReference(1_996_800, 1_996_800, 128_000, conf.TransportReserve)

	for elapsed := 2 * time.Second; elapsed <= 6*time.Second; elapsed += 2 * time.Second {
		decision := state.observe(start.Add(elapsed), utils.TrendDirectionDownward, 600_000, mediumReference, true, conf)
		if elapsed < conf.UnstableDuration && decision.downgrade {
			t.Fatalf("downgraded before unstable duration at %v", elapsed)
		}
		if elapsed == conf.UnstableDuration {
			if !decision.downgrade {
				t.Fatal("sustained insufficient downward estimate did not downgrade")
			}
			state.markDowngrade(start.Add(elapsed), conf)
		}
	}

	for elapsed := 8 * time.Second; elapsed < 18*time.Second; elapsed += 2 * time.Second {
		decision := state.observe(start.Add(elapsed), utils.TrendDirectionUpward, 3_000_000, lowReference, false, conf)
		if decision.upgradeReady {
			t.Fatalf("upgrade became ready before stable duration at %v", elapsed)
		}
	}

	decision := state.observe(start.Add(18*time.Second), utils.TrendDirectionNeutral, 3_000_000, lowReference, false, conf)
	if !decision.upgradeReady {
		t.Fatal("recovered low tier did not become upgrade-ready after stable duration")
	}
	if !estimatedBitrateSupportsUpgrade(3_000_000, mediumReference, conf.UpgradeDiffThreshold) {
		t.Fatal("recovered estimate did not satisfy medium delivery reserve")
	}
	state.markUpgrade(start.Add(18*time.Second), conf)

	decision = state.observe(start.Add(20*time.Second), utils.TrendDirectionNeutral, 3_000_000, mediumReference, false, conf)
	if decision.upgradeReady {
		t.Fatal("second recovery upgrade ignored upgrade backoff")
	}
	decision = state.observe(start.Add(23*time.Second), utils.TrendDirectionNeutral, 3_000_000, mediumReference, false, conf)
	if !decision.upgradeReady {
		t.Fatal("second recovery upgrade was not ready after upgrade backoff")
	}
	if !estimatedBitrateSupportsUpgrade(3_000_000, highReference, conf.UpgradeDiffThreshold) {
		t.Fatal("recovered estimate did not satisfy high delivery reserve")
	}
}

func TestApplicationLimitedTierStartsBoundedOneStepRecoveryProbe(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start, conf)
	lowReference := deliveryBitrateReference(332_800, 332_800, 128_000, conf.TransportReserve)
	mediumReference := deliveryBitrateReference(748_800, 748_800, 128_000, conf.TransportReserve)
	targetBitrate := 650_000

	decision := state.observe(start.Add(conf.RecoveryProbeInterval), utils.TrendDirectionNeutral, targetBitrate, lowReference, false, conf)
	if !decision.upgradeReady {
		t.Fatal("clean tier did not become normally upgrade-ready")
	}
	if state.recoveryProbeReady(start.Add(conf.RecoveryProbeInterval), targetBitrate, lowReference, conf) {
		t.Fatal("tier without a preceding automatic downgrade started a recovery probe")
	}

	// Model a completed downgrade. Clean observations on low become stable, but
	// the application-limited target cannot satisfy the medium-tier threshold.
	start = start.Add(conf.RecoveryProbeInterval)
	state.markDowngrade(start, conf)
	if estimatedBitrateSupportsUpgrade(targetBitrate, mediumReference, conf.UpgradeDiffThreshold) {
		t.Fatal("application-limited target unexpectedly supports the next tier directly")
	}

	for elapsed := 2 * time.Second; elapsed <= conf.RecoveryProbeInterval; elapsed += 2 * time.Second {
		now := start.Add(elapsed)
		decision := state.observe(now, utils.TrendDirectionNeutral, targetBitrate, lowReference, false, conf)
		probeReady := decision.upgradeReady && state.recoveryProbeReady(now, targetBitrate, lowReference, conf)
		if elapsed < conf.RecoveryProbeInterval && probeReady {
			t.Fatalf("recovery probe became ready before its interval at %v", elapsed)
		}
		if elapsed == conf.RecoveryProbeInterval && !probeReady {
			t.Fatal("clean application-limited tier did not become recovery-probe ready")
		}
	}
}

func TestRecoveryProbeRequiresCleanHigherTierBeforeNextStep(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start, conf)
	mediumReference := deliveryBitrateReference(748_800, 748_800, 128_000, conf.TransportReserve)
	targetBitrate := 1_300_000

	state.markDowngrade(start.Add(-conf.RecoveryProbeInterval), conf)
	state.markRecoveryProbe(start, conf)
	for elapsed := 2 * time.Second; elapsed < conf.StableDuration; elapsed += 2 * time.Second {
		state.observe(start.Add(elapsed), utils.TrendDirectionNeutral, targetBitrate, mediumReference, false, conf)
		if state.completeRecoveryProbe(start.Add(elapsed), utils.TrendDirectionNeutral, false, conf) {
			t.Fatalf("recovery probe completed before the stable window at %v", elapsed)
		}
	}

	now := start.Add(conf.StableDuration)
	state.observe(now, utils.TrendDirectionNeutral, targetBitrate, mediumReference, false, conf)
	if !state.completeRecoveryProbe(now, utils.TrendDirectionNeutral, false, conf) {
		t.Fatal("clean recovery probe did not complete at the stable boundary")
	}
	if state.recoveryProbeActive || state.recoveryProbeFailures != 0 {
		t.Fatal("successful recovery probe did not clear its active/failure state")
	}
	if state.recoveryProbeReady(now, targetBitrate, mediumReference, conf) {
		t.Fatal("successful probe allowed another immediate one-tier probe")
	}
}

func TestRecoveryProbeCleanWindowDoesNotDependOnCollapsedTarget(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start.Add(-conf.RecoveryProbeInterval), conf)
	mediumReference := deliveryBitrateReference(748_800, 748_800, 128_000, conf.TransportReserve)

	state.markDowngrade(start.Add(-conf.RecoveryProbeInterval), conf)
	state.markRecoveryProbe(start, conf)

	for elapsed := 2 * time.Second; elapsed <= conf.StableDuration; elapsed += 2 * time.Second {
		now := start.Add(elapsed)
		decision := state.observe(now, utils.TrendDirectionNeutral, 100_000, mediumReference, false, conf)
		if !decision.insufficient {
			t.Fatalf("collapsed target was not classified insufficient at %v", elapsed)
		}
		completed := state.completeRecoveryProbe(now, utils.TrendDirectionNeutral, false, conf)
		if elapsed < conf.StableDuration && completed {
			t.Fatalf("clean probe completed before its window at %v", elapsed)
		}
		if elapsed == conf.StableDuration && !completed {
			t.Fatal("clean probe remained deadlocked on the advisory collapsed target")
		}
	}
}

func TestRecoveryProbeCleanWindowResetsOnReceiverEvidence(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start.Add(-conf.RecoveryProbeInterval), conf)
	mediumReference := deliveryBitrateReference(748_800, 748_800, 128_000, conf.TransportReserve)

	state.markDowngrade(start.Add(-conf.RecoveryProbeInterval), conf)
	state.markRecoveryProbe(start, conf)
	evidenceAt := start.Add(10 * time.Second)
	state.observe(evidenceAt, utils.TrendDirectionNeutral, 1_300_000, mediumReference, true, conf)

	if state.completeRecoveryProbe(start.Add(conf.StableDuration), utils.TrendDirectionNeutral, false, conf) {
		t.Fatal("receiver evidence did not reset the probe's clean stable window")
	}
	cleanAt := evidenceAt.Add(conf.StableDuration)
	state.observe(cleanAt, utils.TrendDirectionNeutral, 1_300_000, mediumReference, false, conf)
	if !state.completeRecoveryProbe(cleanAt, utils.TrendDirectionNeutral, false, conf) {
		t.Fatal("probe did not complete after a full clean interval following receiver evidence")
	}
}

func TestRecoveryProbeDoesNotCountUnobservedTimeAsStable(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start.Add(-conf.RecoveryProbeInterval), conf)
	mediumReference := deliveryBitrateReference(748_800, 748_800, 128_000, conf.TransportReserve)

	state.markDowngrade(start.Add(-conf.RecoveryProbeInterval), conf)
	state.markRecoveryProbe(start, conf)
	state.deferRecoveryProbeObservation(start.Add(2 * conf.StableDuration))

	state.observe(start.Add(2*conf.StableDuration), utils.TrendDirectionNeutral, 1_300_000, mediumReference, false, conf)
	if state.completeRecoveryProbe(start.Add(2*conf.StableDuration), utils.TrendDirectionNeutral, false, conf) {
		t.Fatal("unobserved interval completed the recovery probe")
	}

	now := start.Add(3 * conf.StableDuration)
	state.observe(now, utils.TrendDirectionNeutral, 1_300_000, mediumReference, false, conf)
	if !state.completeRecoveryProbe(now, utils.TrendDirectionNeutral, false, conf) {
		t.Fatal("probe did not complete after a fresh observed stable window")
	}
}

func TestFailedRecoveryProbeUsesExponentialBoundedBackoff(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start, conf)
	lowReference := deliveryBitrateReference(332_800, 332_800, 128_000, conf.TransportReserve)
	targetBitrate := 650_000

	state.markDowngrade(start.Add(-conf.RecoveryProbeInterval), conf)
	state.markRecoveryProbe(start, conf)
	failed, backoff := state.markDowngrade(start.Add(conf.UnstableDuration), conf)
	if !failed || backoff != conf.RecoveryProbeInterval {
		t.Fatalf("first failed probe = (%v, %v), want (true, %v)", failed, backoff, conf.RecoveryProbeInterval)
	}

	firstRetry := start.Add(conf.UnstableDuration + conf.RecoveryProbeInterval)
	decision := state.observe(firstRetry, utils.TrendDirectionNeutral, targetBitrate, lowReference, false, conf)
	if !decision.upgradeReady || !state.recoveryProbeReady(firstRetry, targetBitrate, lowReference, conf) {
		t.Fatal("first failed probe did not reopen at its bounded retry time")
	}

	state.markRecoveryProbe(firstRetry, conf)
	secondFailure := firstRetry.Add(conf.UnstableDuration)
	failed, backoff = state.markDowngrade(secondFailure, conf)
	if !failed || backoff != 2*conf.RecoveryProbeInterval {
		t.Fatalf("second failed probe = (%v, %v), want (true, %v)", failed, backoff, 2*conf.RecoveryProbeInterval)
	}

	decision = state.observe(secondFailure.Add(conf.StableDuration), utils.TrendDirectionNeutral, targetBitrate, lowReference, false, conf)
	if !decision.upgradeReady {
		t.Fatal("lower tier did not become clean while probe retry remained backed off")
	}
	if !state.upgradeBlockedByRecoveryBackoff(secondFailure.Add(2*conf.RecoveryProbeInterval - time.Second)) {
		t.Fatal("failed probe did not block an otherwise capacity-qualified upgrade")
	}
	if state.recoveryProbeReady(secondFailure.Add(2*conf.RecoveryProbeInterval-time.Second), targetBitrate, lowReference, conf) {
		t.Fatal("failed-probe backoff expired early")
	}
	if state.upgradeBlockedByRecoveryBackoff(secondFailure.Add(2 * conf.RecoveryProbeInterval)) {
		t.Fatal("failed-probe upgrade backoff did not expire at its boundary")
	}
	if !state.recoveryProbeReady(secondFailure.Add(2*conf.RecoveryProbeInterval), targetBitrate, lowReference, conf) {
		t.Fatal("failed-probe backoff did not expire at its boundary")
	}
	state.markUpgrade(secondFailure.Add(2*conf.RecoveryProbeInterval), conf)
	if state.recoveryProbeFailures != 0 {
		t.Fatal("capacity-qualified recovery did not clear failed-probe backoff")
	}

	if got, want := recoveryProbeBackoff(10, conf), conf.RecoveryProbeMaxBackoff; got != want {
		t.Fatalf("capped recovery backoff = %v, want %v", got, want)
	}
}

func TestRecoveryProbeStateRemainsPeerLocal(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	healthy := newEstimatorObservationState(start, conf)
	constrained := newEstimatorObservationState(start, conf)

	constrained.markDowngrade(start.Add(-conf.RecoveryProbeInterval), conf)
	constrained.markRecoveryProbe(start, conf)
	constrained.markDowngrade(start.Add(conf.UnstableDuration), conf)

	if healthy.recoveryProbeActive || healthy.recoveryProbeFailures != 0 || healthy.recoverySteps != 0 {
		t.Fatal("constrained peer recovery probe changed healthy peer state")
	}
	if constrained.recoveryProbeActive || constrained.recoveryProbeFailures != 1 {
		t.Fatal("constrained peer did not retain its own failed-probe state")
	}
}

func TestApplicationLimitedRecoveryRestoresTwoTiersInOrder(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start.Add(-3*conf.RecoveryProbeInterval), conf)
	lowReference := deliveryBitrateReference(332_800, 332_800, 128_000, conf.TransportReserve)
	mediumReference := deliveryBitrateReference(748_800, 748_800, 128_000, conf.TransportReserve)

	state.markDowngrade(start.Add(-2*conf.RecoveryProbeInterval), conf)
	state.markDowngrade(start.Add(-conf.RecoveryProbeInterval), conf)
	if state.recoverySteps != 2 {
		t.Fatalf("recovery steps = %d, want 2", state.recoverySteps)
	}

	decision := state.observe(start, utils.TrendDirectionNeutral, 650_000, lowReference, false, conf)
	if !decision.upgradeReady || !state.recoveryProbeReady(start, 650_000, lowReference, conf) {
		t.Fatal("clean low tier did not become ready for its first recovery step")
	}
	state.markRecoveryProbe(start, conf)
	if state.recoverySteps != 1 {
		t.Fatalf("recovery steps after low-to-medium probe = %d, want 1", state.recoverySteps)
	}

	mediumStable := start.Add(conf.StableDuration)
	state.observe(mediumStable, utils.TrendDirectionNeutral, 1_300_000, mediumReference, false, conf)
	if !state.completeRecoveryProbe(mediumStable, utils.TrendDirectionNeutral, false, conf) {
		t.Fatal("clean medium tier did not complete its recovery probe")
	}
	if state.recoveryProbeReady(mediumStable, 1_300_000, mediumReference, conf) {
		t.Fatal("medium-to-high recovery probe started without its own interval")
	}

	highProbeAt := mediumStable.Add(conf.RecoveryProbeInterval)
	decision = state.observe(highProbeAt, utils.TrendDirectionNeutral, 1_300_000, mediumReference, false, conf)
	if !decision.upgradeReady || !state.recoveryProbeReady(highProbeAt, 1_300_000, mediumReference, conf) {
		t.Fatal("clean medium tier did not become ready for its second recovery step")
	}
	state.markRecoveryProbe(highProbeAt, conf)
	if state.recoverySteps != 0 {
		t.Fatalf("recovery steps after medium-to-high probe = %d, want 0", state.recoverySteps)
	}
}

func TestEstimatorHysteresisPreventsRapidOscillation(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	state := newEstimatorObservationState(start, conf)
	currentReference := deliveryBitrateReference(748_800, 748_800, 128_000, conf.TransportReserve)
	upgradeReference := deliveryBitrateReference(1_996_800, 0, 128_000, conf.TransportReserve)
	targetBitrate := 850_000

	if estimatedBitrateRequiresDowngrade(targetBitrate, currentReference, conf.DowngradeDeficitThreshold) {
		t.Fatal("deadband estimate unexpectedly requires downgrade")
	}
	if estimatedBitrateSupportsUpgrade(targetBitrate, upgradeReference, conf.UpgradeDiffThreshold) {
		t.Fatal("deadband estimate unexpectedly supports upgrade")
	}

	for elapsed := 2 * time.Second; elapsed <= 60*time.Second; elapsed += 2 * time.Second {
		decision := state.observe(start.Add(elapsed), utils.TrendDirectionNeutral, targetBitrate, currentReference, false, conf)
		if decision.downgrade {
			t.Fatalf("deadband estimate caused downgrade at %v", elapsed)
		}
		if decision.upgradeReady && estimatedBitrateSupportsUpgrade(targetBitrate, upgradeReference, conf.UpgradeDiffThreshold) {
			t.Fatalf("deadband estimate caused upgrade at %v", elapsed)
		}
	}
}

func TestEstimatorStartupAndBackoffWindowsRemainBounded(t *testing.T) {
	conf := adaptiveEstimatorTestConfig()
	start := time.Unix(1_700_000_000, 0)
	reference := deliveryBitrateReference(1_996_800, 1_996_800, 128_000, conf.TransportReserve)

	state := newEstimatorObservationState(start, conf)
	decision := state.observe(start.Add(conf.StalledDuration), utils.TrendDirectionNeutral, 1_300_000, reference, true, conf)
	if decision.stalled || decision.downgrade {
		t.Fatal("stalled window expired at its boundary instead of after it")
	}

	state = newEstimatorObservationState(start, conf)
	for elapsed := 2 * time.Second; elapsed <= conf.UnstableDuration; elapsed += 2 * time.Second {
		decision = state.observe(start.Add(elapsed), utils.TrendDirectionDownward, 1_300_000, reference, true, conf)
	}
	if !decision.downgrade {
		t.Fatal("downgrade was not ready at the unchanged unstable-duration boundary")
	}
	state.markDowngrade(start.Add(conf.UnstableDuration), conf)

	decision = state.observe(start.Add(conf.UnstableDuration+conf.DowngradeBackoff-time.Second), utils.TrendDirectionDownward, 1_300_000, reference, true, conf)
	if decision.downgrade {
		t.Fatal("downgrade backoff expired early")
	}
	decision = state.observe(start.Add(conf.UnstableDuration+conf.DowngradeBackoff), utils.TrendDirectionDownward, 1_300_000, reference, true, conf)
	if !decision.downgrade {
		t.Fatal("downgrade backoff did not expire at the configured boundary")
	}

	state = newEstimatorObservationState(start, conf)
	state.markUpgrade(start.Add(conf.StableDuration), conf)
	decision = state.observe(start.Add(conf.StableDuration+conf.UpgradeBackoff-time.Second), utils.TrendDirectionNeutral, 3_000_000, reference, false, conf)
	if decision.upgradeReady {
		t.Fatal("upgrade backoff expired early")
	}
	decision = state.observe(start.Add(conf.StableDuration+conf.UpgradeBackoff), utils.TrendDirectionNeutral, 3_000_000, reference, false, conf)
	if !decision.upgradeReady {
		t.Fatal("upgrade backoff did not expire at the configured boundary")
	}
}

func TestDeliveryBitrateReferenceUsesNominalMeasuredAudioAndTransport(t *testing.T) {
	if got, want := deliveryBitrateReference(1_900_000, 1_996_800, 128_000, 0.05), uint64(2_231_040); got != want {
		t.Fatalf("nominal delivery reference = %d, want %d", got, want)
	}
	if got, want := deliveryBitrateReference(2_100_000, 1_996_800, 128_000, 0.05), uint64(2_339_400); got != want {
		t.Fatalf("measured delivery reference = %d, want %d", got, want)
	}
}

func TestEstimatedBitrateRequiresDowngrade(t *testing.T) {
	reference := uint64(2_231_040)
	if estimatedBitrateRequiresDowngrade(2_000_000, reference, 0.15) {
		t.Fatal("estimate inside tolerated deficit unexpectedly requires downgrade")
	}
	if !estimatedBitrateRequiresDowngrade(1_800_000, reference, 0.15) {
		t.Fatal("materially insufficient estimate did not require downgrade")
	}
}

func TestEstimatedBitrateSupportsUpgrade(t *testing.T) {
	tests := []struct {
		name             string
		targetBitrate    int
		referenceBitrate uint64
		threshold        float64
		want             bool
	}{
		{
			name:             "default threshold accepts more than fifteen percent headroom",
			targetBitrate:    1_160_000,
			referenceBitrate: 1_000_000,
			threshold:        0.15,
			want:             true,
		},
		{
			name:             "configured medium tier rejects insufficient estimate",
			targetBitrate:    613_076,
			referenceBitrate: 748_800,
			threshold:        0.15,
			want:             false,
		},
		{
			name:             "configured medium tier accepts sufficient estimate",
			targetBitrate:    950_000,
			referenceBitrate: 748_800,
			threshold:        0.15,
			want:             true,
		},
		{
			name:             "negative threshold disables the check",
			targetBitrate:    0,
			referenceBitrate: 1_000_000,
			threshold:        -1,
			want:             true,
		},
		{
			name:             "zero reference bitrate is not actionable",
			targetBitrate:    1_000_000,
			referenceBitrate: 0,
			threshold:        0.15,
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := estimatedBitrateSupportsUpgrade(tt.targetBitrate, tt.referenceBitrate, tt.threshold); got != tt.want {
				t.Fatalf("estimatedBitrateSupportsUpgrade(%d, %d, %v) = %v, want %v", tt.targetBitrate, tt.referenceBitrate, tt.threshold, got, tt.want)
			}
		})
	}
}

func TestReferenceBitrateForUpgrade(t *testing.T) {
	if got, want := referenceBitrateForUpgrade(200_000, 332_800), uint64(332_800); got != want {
		t.Fatalf("nominal reference = %d, want %d", got, want)
	}
	if got, want := referenceBitrateForUpgrade(200_000, 0), uint64(200_000); got != want {
		t.Fatalf("measured fallback = %d, want %d", got, want)
	}
}
