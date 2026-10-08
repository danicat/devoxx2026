package audio

import (
	"fmt"
	"math"
	"sync"

	ebitenaudio "github.com/hajimehoshi/ebiten/v2/audio"
)

// AudioManager defines the required public interface for game sound effects.
type AudioManager interface {
	PlaySFXSweep()
	PlaySFXAllocate()
	PlaySFXDeallocate()
	PlaySFXPromote()
	PlaySFXSTW()
	PlaySFXOOMAlarm()
	Update()
}

// Default cooldown frames at 60 TPS to prevent audio clipping/distortion.
const (
	CooldownSweepFrames      = 3  // ~50ms
	CooldownAllocateFrames   = 4  // ~66ms
	CooldownDeallocateFrames = 2  // ~33ms
	CooldownPromoteFrames    = 6  // ~100ms
	CooldownSTWFrames        = 30 // ~500ms
	CooldownOOMAlarmFrames   = 18 // ~300ms
)

// audioManagerImpl implements AudioManager.
type audioManagerImpl struct {
	context *ebitenaudio.Context

	poolSweep      *VoicePool
	poolAllocate   *VoicePool
	poolDeallocate *VoicePool
	poolPromote    *VoicePool
	poolSTW        *VoicePool
	poolOOMAlarm   *VoicePool

	cooldownSweep      int
	cooldownAllocate   int
	cooldownDeallocate int
	cooldownPromote    int
	cooldownSTW        int
	cooldownOOMAlarm   int

	mu sync.Mutex
}

// NewAudioManager initializes the audio system with the standard 44.1kHz sample rate.
// Returns an error if the underlying audio context cannot be initialized.
func NewAudioManager() (AudioManager, error) {
	ctx, err := GetOrCreateAudioContext(SampleRate)
	if err != nil {
		return nil, fmt.Errorf("audio manager initialization failed: %w", err)
	}
	return NewAudioManagerWithContext(ctx)
}

// NewHeadlessAudioManager creates a mock-free silent AudioManager for testing and headless environments.
// It verifies DSP sample generation and handles all SFX triggers and rate limiting without hardware.
func NewHeadlessAudioManager() AudioManager {
	mgr, _ := NewAudioManagerWithContext(nil)
	return mgr
}

// NewAudioManagerWithContext creates an AudioManager with the provided ebiten audio context.
// If ctx is nil, the manager operates in silent/headless mode.
func NewAudioManagerWithContext(ctx *ebitenaudio.Context) (AudioManager, error) {
	mgr := &audioManagerImpl{
		context: ctx,
	}

	// 1. Synthesize Sweep PCM (FM synthesis with resonant decay)
	sweepPCM := buildSweepPCM()
	poolSweep, err := NewVoicePool(ctx, sweepPCM, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to create sweep voice pool: %w", err)
	}
	mgr.poolSweep = poolSweep

	// 2. Synthesize Allocate PCM (High-pitch square-wave bleeps)
	allocatePCM := buildAllocatePCM()
	poolAllocate, err := NewVoicePool(ctx, allocatePCM, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to create allocate voice pool: %w", err)
	}
	mgr.poolAllocate = poolAllocate

	// 3. Synthesize Deallocate PCM (Noise burst into soft click/pop)
	deallocatePCM := buildDeallocatePCM()
	poolDeallocate, err := NewVoicePool(ctx, deallocatePCM, 6)
	if err != nil {
		return nil, fmt.Errorf("failed to create deallocate voice pool: %w", err)
	}
	mgr.poolDeallocate = poolDeallocate

	// 4. Synthesize Promote PCM (Ascending dual-tone arpeggio)
	promotePCM := buildPromotePCM()
	poolPromote, err := NewVoicePool(ctx, promotePCM, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to create promote voice pool: %w", err)
	}
	mgr.poolPromote = poolPromote

	// 5. Synthesize STW PCM (Deep low-pass resonant rumble/hum)
	stwPCM := buildSTWPCM()
	poolSTW, err := NewVoicePool(ctx, stwPCM, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to create STW voice pool: %w", err)
	}
	mgr.poolSTW = poolSTW

	// 6. Synthesize OOM Alarm PCM (Emergency klaxon warning)
	oomPCM := buildOOMAlarmPCM()
	poolOOM, err := NewVoicePool(ctx, oomPCM, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to create OOM alarm voice pool: %w", err)
	}
	mgr.poolOOMAlarm = poolOOM

	return mgr, nil
}

// Update advances internal cooldown timers. Should be called once per frame (60 TPS).
func (m *audioManagerImpl) Update() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cooldownSweep > 0 {
		m.cooldownSweep--
	}
	if m.cooldownAllocate > 0 {
		m.cooldownAllocate--
	}
	if m.cooldownDeallocate > 0 {
		m.cooldownDeallocate--
	}
	if m.cooldownPromote > 0 {
		m.cooldownPromote--
	}
	if m.cooldownSTW > 0 {
		m.cooldownSTW--
	}
	if m.cooldownOOMAlarm > 0 {
		m.cooldownOOMAlarm--
	}
}

// PlaySFXSweep plays the FM laser sweep sound if cooldown has elapsed.
func (m *audioManagerImpl) PlaySFXSweep() {
	m.mu.Lock()
	if m.cooldownSweep > 0 {
		m.mu.Unlock()
		return
	}
	m.cooldownSweep = CooldownSweepFrames
	pool := m.poolSweep
	m.mu.Unlock()

	pool.Play()
}

// PlaySFXAllocate plays the object spawn square-wave bleep if cooldown has elapsed.
func (m *audioManagerImpl) PlaySFXAllocate() {
	m.mu.Lock()
	if m.cooldownAllocate > 0 {
		m.mu.Unlock()
		return
	}
	m.cooldownAllocate = CooldownAllocateFrames
	pool := m.poolAllocate
	m.mu.Unlock()

	pool.Play()
}

// PlaySFXDeallocate plays the noise pop deallocation sound if cooldown has elapsed.
func (m *audioManagerImpl) PlaySFXDeallocate() {
	m.mu.Lock()
	if m.cooldownDeallocate > 0 {
		m.mu.Unlock()
		return
	}
	m.cooldownDeallocate = CooldownDeallocateFrames
	pool := m.poolDeallocate
	m.mu.Unlock()

	pool.Play()
}

// PlaySFXPromote plays the ascending promotion arpeggio chime if cooldown has elapsed.
func (m *audioManagerImpl) PlaySFXPromote() {
	m.mu.Lock()
	if m.cooldownPromote > 0 {
		m.mu.Unlock()
		return
	}
	m.cooldownPromote = CooldownPromoteFrames
	pool := m.poolPromote
	m.mu.Unlock()

	pool.Play()
}

// PlaySFXSTW plays the deep Stop-The-World pause resonant rumble if cooldown has elapsed.
func (m *audioManagerImpl) PlaySFXSTW() {
	m.mu.Lock()
	if m.cooldownSTW > 0 {
		m.mu.Unlock()
		return
	}
	m.cooldownSTW = CooldownSTWFrames
	pool := m.poolSTW
	m.mu.Unlock()

	pool.Play()
}

// PlaySFXOOMAlarm plays the emergency Old Gen klaxon alarm if cooldown has elapsed.
func (m *audioManagerImpl) PlaySFXOOMAlarm() {
	m.mu.Lock()
	if m.cooldownOOMAlarm > 0 {
		m.mu.Unlock()
		return
	}
	m.cooldownOOMAlarm = CooldownOOMAlarmFrames
	pool := m.poolOOMAlarm
	m.mu.Unlock()

	pool.Play()
}

// --- Procedural Sound Synthesis Builders ---

// buildSweepPCM generates an FM synthesis laser collection sweep.
// Sweeps from 880 Hz down to 220 Hz with decaying modulation.
func buildSweepPCM() []byte {
	duration := 0.16
	carrierFn := func(t float64) float64 {
		progress := t / duration
		// Exponential-like downward sweep
		return 880.0 - 660.0*math.Pow(progress, 0.7)
	}
	modFn := func(t float64) float64 {
		progress := t / duration
		return 220.0 - 100.0*progress
	}
	modIndexFn := func(t float64) float64 {
		progress := t / duration
		return 3.5 * (1.0 - progress)
	}
	env := ADSREnvelope{
		AttackDuration:  0.003,
		DecayDuration:   0.060,
		SustainLevel:    0.25,
		ReleaseDuration: 0.080,
	}

	samples := GenerateFMSweep(SampleRate, duration, carrierFn, modFn, modIndexFn, env)
	return FloatsToStereoPCM16(samples)
}

// buildAllocatePCM generates a high-pitched 8-bit square-wave allocation chirp.
func buildAllocatePCM() []byte {
	duration := 0.065
	numSamples := int(float64(SampleRate) * duration)
	samples := make([]float64, numSamples)

	env := ADSREnvelope{
		AttackDuration:  0.002,
		DecayDuration:   0.020,
		SustainLevel:    0.60,
		ReleaseDuration: 0.040,
	}

	var phase float64
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(SampleRate)
		progress := t / duration

		var freq float64
		var duty float64
		if progress < 0.5 {
			freq = 1174.66 // D6
			duty = 0.25
		} else {
			freq = 1567.98 // G6
			duty = 0.50
		}

		amp := env.ValueAt(t, duration) * 0.7
		samples[i] = Square(phase, duty) * amp

		phase += 2.0 * math.Pi * freq / float64(SampleRate)
		if phase >= 2.0*math.Pi {
			phase = math.Mod(phase, 2.0*math.Pi)
		}
	}

	return FloatsToStereoPCM16(samples)
}

// buildDeallocatePCM generates a soft noise burst transitioning into a low pop.
func buildDeallocatePCM() []byte {
	duration := 0.075
	numSamples := int(float64(SampleRate) * duration)
	rawSamples := make([]float64, numSamples)

	envNoise := ADSREnvelope{
		AttackDuration:  0.001,
		DecayDuration:   0.020,
		SustainLevel:    0.0,
		ReleaseDuration: 0.005,
	}

	envPop := ADSREnvelope{
		AttackDuration:  0.002,
		DecayDuration:   0.035,
		SustainLevel:    0.15,
		ReleaseDuration: 0.030,
	}

	var popPhase float64
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(SampleRate)
		progress := t / duration

		// Noise component
		noiseAmp := envNoise.ValueAt(t, duration) * 0.5
		noiseVal := Noise() * noiseAmp

		// Pop sine pitch drops rapidly from 280Hz to 60Hz
		popFreq := 280.0 - 220.0*math.Sqrt(progress)
		popAmp := envPop.ValueAt(t, duration) * 0.6
		popVal := Sine(popPhase) * popAmp

		rawSamples[i] = noiseVal + popVal

		popPhase += 2.0 * math.Pi * popFreq / float64(SampleRate)
		if popPhase >= 2.0*math.Pi {
			popPhase = math.Mod(popPhase, 2.0*math.Pi)
		}
	}

	// Filter harsh high frequencies for a tactile, soft bubble pop sound
	filtered := ApplyLowPassFilter(rawSamples, SampleRate, 1400.0)
	return FloatsToStereoPCM16(filtered)
}

// buildPromotePCM generates an ascending dual-tone/tri-tone arpeggio.
func buildPromotePCM() []byte {
	duration := 0.18
	numSamples := int(float64(SampleRate) * duration)
	samples := make([]float64, numSamples)

	noteDur := duration / 3.0
	noteEnv := ADSREnvelope{
		AttackDuration:  0.003,
		DecayDuration:   0.025,
		SustainLevel:    0.70,
		ReleaseDuration: 0.030,
	}

	var phase float64
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(SampleRate)
		noteIdx := int(t / noteDur)
		if noteIdx > 2 {
			noteIdx = 2
		}

		tInNote := math.Mod(t, noteDur)

		var freq float64
		switch noteIdx {
		case 0:
			freq = 587.33 // D5
		case 1:
			freq = 739.99 // F#5
		case 2:
			freq = 880.00 // A5
		}

		amp := noteEnv.ValueAt(tInNote, noteDur) * 0.65
		// Triangle blended with sine for smooth harmonic chiptune chime
		wave := 0.6*Triangle(phase) + 0.4*Sine(phase)
		samples[i] = wave * amp

		phase += 2.0 * math.Pi * freq / float64(SampleRate)
		if phase >= 2.0*math.Pi {
			phase = math.Mod(phase, 2.0*math.Pi)
		}
	}

	return FloatsToStereoPCM16(samples)
}

// buildSTWPCM generates a deep low-pass resonant rumble/hum for Stop-The-World pauses.
func buildSTWPCM() []byte {
	duration := 0.65
	numSamples := int(float64(SampleRate) * duration)
	rawSamples := make([]float64, numSamples)

	env := ADSREnvelope{
		AttackDuration:  0.05,
		DecayDuration:   0.15,
		SustainLevel:    0.55,
		ReleaseDuration: 0.40,
	}

	var subPhase float64
	var fundPhase float64
	var sawPhase float64

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(SampleRate)

		// 6.5 Hz tremolo LFO to create a pulsating engine hum
		lfo := 0.75 + 0.25*math.Sin(2.0*math.Pi*6.5*t)
		amp := env.ValueAt(t, duration) * lfo * 0.8

		// Layered sub-bass frequencies: 55Hz fundamental, 110Hz second harmonic, 27.5Hz sub
		subVal := 0.5 * Sine(subPhase)
		fundVal := 0.3 * Sine(fundPhase)
		sawVal := 0.2 * Sawtooth(sawPhase)

		rawSamples[i] = (subVal + fundVal + sawVal) * amp

		subPhase += 2.0 * math.Pi * 55.0 / float64(SampleRate)
		fundPhase += 2.0 * math.Pi * 110.0 / float64(SampleRate)
		sawPhase += 2.0 * math.Pi * 55.0 / float64(SampleRate)

		if subPhase >= 2.0*math.Pi {
			subPhase = math.Mod(subPhase, 2.0*math.Pi)
		}
		if fundPhase >= 2.0*math.Pi {
			fundPhase = math.Mod(fundPhase, 2.0*math.Pi)
		}
		if sawPhase >= 2.0*math.Pi {
			sawPhase = math.Mod(sawPhase, 2.0*math.Pi)
		}
	}

	// Low-pass filter at 220Hz for a deep, resonant rumble
	filtered := ApplyLowPassFilter(rawSamples, SampleRate, 220.0)
	return FloatsToStereoPCM16(filtered)
}

// buildOOMAlarmPCM generates an alternating emergency warning klaxon for critical Old Gen capacity.
func buildOOMAlarmPCM() []byte {
	duration := 0.32
	numSamples := int(float64(SampleRate) * duration)
	samples := make([]float64, numSamples)

	env := ADSREnvelope{
		AttackDuration:  0.005,
		DecayDuration:   0.030,
		SustainLevel:    0.80,
		ReleaseDuration: 0.040,
	}

	var phase float64
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(SampleRate)
		progress := t / duration

		var freq float64
		if progress < 0.5 {
			freq = 880.0 // A5 high alarm tone
		} else {
			freq = 659.25 // E5 lower alarm tone
		}

		amp := env.ValueAt(t, duration) * 0.7
		// Square wave with 35% duty cycle plus slight saw grit
		wave := 0.8*Square(phase, 0.35) + 0.2*Sawtooth(phase)
		samples[i] = wave * amp

		phase += 2.0 * math.Pi * freq / float64(SampleRate)
		if phase >= 2.0*math.Pi {
			phase = math.Mod(phase, 2.0*math.Pi)
		}
	}

	return FloatsToStereoPCM16(samples)
}
