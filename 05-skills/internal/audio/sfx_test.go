package audio

import (
	"testing"
)

func TestPCMBuilders(t *testing.T) {
	builders := []struct {
		name string
		fn   func() []byte
	}{
		{"Sweep", buildSweepPCM},
		{"Allocate", buildAllocatePCM},
		{"Deallocate", buildDeallocatePCM},
		{"Promote", buildPromotePCM},
		{"STW", buildSTWPCM},
		{"OOMAlarm", buildOOMAlarmPCM},
	}

	for _, b := range builders {
		t.Run(b.name, func(t *testing.T) {
			pcm := b.fn()
			if len(pcm) == 0 {
				t.Fatalf("%s returned empty PCM buffer", b.name)
			}
			if len(pcm)%BytesPerSampleFrame != 0 {
				t.Fatalf("%s length %d is not aligned to BytesPerSampleFrame (%d)", b.name, len(pcm), BytesPerSampleFrame)
			}

			// Ensure audio is not pure silence (verify at least some non-zero samples exist)
			hasNonZero := false
			for _, bVal := range pcm {
				if bVal != 0 {
					hasNonZero = true
					break
				}
			}
			if !hasNonZero {
				t.Errorf("%s PCM is completely silent (all zeros)", b.name)
			}
		})
	}
}

func TestHeadlessAudioManagerLifecycle(t *testing.T) {
	mgr := NewHeadlessAudioManager()
	if mgr == nil {
		t.Fatal("NewHeadlessAudioManager returned nil")
	}

	// Trigger all SFX without error or panic
	mgr.PlaySFXSweep()
	mgr.PlaySFXAllocate()
	mgr.PlaySFXDeallocate()
	mgr.PlaySFXPromote()
	mgr.PlaySFXSTW()
	mgr.PlaySFXOOMAlarm()

	// Update cycle
	for i := 0; i < 60; i++ {
		mgr.Update()
	}
}

func TestSFXRateLimitingAndCooldowns(t *testing.T) {
	mgrImpl, err := NewAudioManagerWithContext(nil)
	if err != nil {
		t.Fatalf("Failed to create audio manager: %v", err)
	}

	impl, ok := mgrImpl.(*audioManagerImpl)
	if !ok {
		t.Fatalf("Expected *audioManagerImpl type")
	}

	// 1. Test SFXSweep cooldown
	if impl.cooldownSweep != 0 {
		t.Errorf("Initial cooldownSweep = %d, expected 0", impl.cooldownSweep)
	}
	mgrImpl.PlaySFXSweep()
	if impl.cooldownSweep != CooldownSweepFrames {
		t.Errorf("After PlaySFXSweep(), cooldownSweep = %d, expected %d", impl.cooldownSweep, CooldownSweepFrames)
	}

	// Second play call while on cooldown should keep existing cooldown
	mgrImpl.PlaySFXSweep()
	if impl.cooldownSweep != CooldownSweepFrames {
		t.Errorf("Cooldown changed unexpectedly during cooldown period: %d", impl.cooldownSweep)
	}

	// Decrement via Update
	mgrImpl.Update()
	if impl.cooldownSweep != CooldownSweepFrames-1 {
		t.Errorf("After 1 Update(), cooldownSweep = %d, expected %d", impl.cooldownSweep, CooldownSweepFrames-1)
	}

	for i := 0; i < CooldownSweepFrames-1; i++ {
		mgrImpl.Update()
	}
	if impl.cooldownSweep != 0 {
		t.Errorf("After full cooldown period, cooldownSweep = %d, expected 0", impl.cooldownSweep)
	}

	// 2. Test SFXDeallocate cooldown
	mgrImpl.PlaySFXDeallocate()
	if impl.cooldownDeallocate != CooldownDeallocateFrames {
		t.Errorf("cooldownDeallocate = %d, expected %d", impl.cooldownDeallocate, CooldownDeallocateFrames)
	}
	for i := 0; i < CooldownDeallocateFrames; i++ {
		mgrImpl.Update()
	}
	if impl.cooldownDeallocate != 0 {
		t.Errorf("cooldownDeallocate = %d, expected 0 after cooldown period", impl.cooldownDeallocate)
	}

	// 3. Test SFXSTW cooldown
	mgrImpl.PlaySFXSTW()
	if impl.cooldownSTW != CooldownSTWFrames {
		t.Errorf("cooldownSTW = %d, expected %d", impl.cooldownSTW, CooldownSTWFrames)
	}
	for i := 0; i < CooldownSTWFrames; i++ {
		mgrImpl.Update()
	}
	if impl.cooldownSTW != 0 {
		t.Errorf("cooldownSTW = %d, expected 0 after cooldown period", impl.cooldownSTW)
	}
}

func TestLiveAudioManagerCreation(t *testing.T) {
	// Tests NewAudioManager() initialization.
	// If the host platform supports audio context, it must succeed without error.
	// If audio context fails (e.g. headless container without audio hardware), it must fail-fast with error, not panic.
	mgr, err := NewAudioManager()
	if err != nil {
		t.Logf("NewAudioManager() failed with error (expected in some headless/driverless environments): %v", err)
		return
	}
	if mgr == nil {
		t.Fatal("NewAudioManager() returned nil manager with no error")
	}

	// Verify all triggers operate cleanly with live players
	mgr.PlaySFXSweep()
	mgr.PlaySFXAllocate()
	mgr.PlaySFXDeallocate()
	mgr.PlaySFXPromote()
	mgr.PlaySFXSTW()
	mgr.PlaySFXOOMAlarm()
	mgr.Update()
}
