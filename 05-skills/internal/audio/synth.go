package audio

import (
	"errors"
	"fmt"
	"math"
	"sync"

	ebitenaudio "github.com/hajimehoshi/ebiten/v2/audio"
)

// Audio system constants.
const (
	SampleRate          = 44100
	BytesPerSampleFrame = 4 // 16-bit stereo: 2 channels * 2 bytes
)

// WaveType represents the basic procedural oscillator waveform type.
type WaveType int

const (
	WaveSine WaveType = iota
	WaveSquare
	WaveSawtooth
	WaveTriangle
	WaveNoise
)

// ADSREnvelope represents an Attack, Decay, Sustain, Release volume envelope.
// Durations are in seconds, SustainLevel is in [0.0, 1.0].
type ADSREnvelope struct {
	AttackDuration  float64
	DecayDuration   float64
	SustainLevel    float64
	ReleaseDuration float64
}

// ValueAt evaluates the ADSR envelope amplitude at time t (in seconds)
// given the sound's total duration in seconds.
func (e ADSREnvelope) ValueAt(t, totalDuration float64) float64 {
	if totalDuration <= 0 || t < 0 || t >= totalDuration {
		return 0.0
	}

	a := math.Max(0, e.AttackDuration)
	d := math.Max(0, e.DecayDuration)
	r := math.Max(0, e.ReleaseDuration)
	s := math.Max(0.0, math.Min(1.0, e.SustainLevel))

	// Scale envelope phases if they exceed total duration.
	sum := a + d + r
	if sum > totalDuration && sum > 0 {
		scale := totalDuration / sum
		a *= scale
		d *= scale
		r *= scale
	}

	sustainStart := a + d
	releaseStart := totalDuration - r
	if releaseStart < sustainStart {
		releaseStart = sustainStart
	}

	switch {
	case a > 0 && t < a:
		// Attack phase: linear ramp from 0.0 to 1.0
		return t / a
	case d > 0 && t < sustainStart:
		// Decay phase: linear ramp from 1.0 down to sustain level
		decayProgress := (t - a) / d
		return 1.0 - (1.0-s)*decayProgress
	case t < releaseStart:
		// Sustain phase
		return s
	default:
		// Release phase: ramp from sustain level down to 0.0
		if r <= 0 {
			return 0.0
		}
		relProgress := (t - releaseStart) / r
		if relProgress >= 1.0 {
			return 0.0
		}
		return s * (1.0 - relProgress)
	}
}

// Sine returns a sine wave sample in [-1.0, 1.0] for the given phase in radians.
func Sine(phase float64) float64 {
	return math.Sin(phase)
}

// Square returns a square wave sample with variable duty cycle in (0.0, 1.0).
func Square(phase float64, dutyCycle float64) float64 {
	if dutyCycle <= 0.0 || dutyCycle >= 1.0 {
		dutyCycle = 0.5
	}
	norm := phase / (2.0 * math.Pi)
	norm = norm - math.Floor(norm)
	if norm < dutyCycle {
		return 1.0
	}
	return -1.0
}

// Sawtooth returns a sawtooth wave sample in [-1.0, 1.0].
func Sawtooth(phase float64) float64 {
	norm := phase / (2.0 * math.Pi)
	norm = norm - math.Floor(norm)
	return 2.0*norm - 1.0
}

// Triangle returns a triangle wave sample in [-1.0, 1.0].
func Triangle(phase float64) float64 {
	norm := phase / (2.0 * math.Pi)
	norm = norm - math.Floor(norm)
	return 4.0*math.Abs(norm-0.5) - 1.0
}

// Fast deterministic 64-bit LCG/XorShift state for procedural white noise.
var (
	noiseState uint64 = 0x853c49e6748fea9b
	noiseMu    sync.Mutex
)

// Noise returns a white noise sample in [-1.0, 1.0].
func Noise() float64 {
	noiseMu.Lock()
	noiseState ^= noiseState << 13
	noiseState ^= noiseState >> 7
	noiseState ^= noiseState << 17
	val := noiseState
	noiseMu.Unlock()

	// Convert 64-bit int to float64 in [-1.0, 1.0]
	return (float64(val&0x1FFFFFFFFFFFFF)/float64(0x1FFFFFFFFFFFFF))*2.0 - 1.0
}

// DeterministicNoise generates a noise sample with explicit state.
func DeterministicNoise(state *uint64) float64 {
	*state ^= *state << 13
	*state ^= *state >> 7
	*state ^= *state << 17
	val := *state
	return (float64(val&0x1FFFFFFFFFFFFF)/float64(0x1FFFFFFFFFFFFF))*2.0 - 1.0
}

// ApplyLowPassFilter applies a single-pole low-pass filter to an array of samples.
// cutoffHz is the filter cutoff frequency in Hz.
func ApplyLowPassFilter(samples []float64, sampleRate int, cutoffHz float64) []float64 {
	if len(samples) == 0 || cutoffHz <= 0 {
		return samples
	}
	dt := 1.0 / float64(sampleRate)
	rc := 1.0 / (2.0 * math.Pi * cutoffHz)
	alpha := dt / (rc + dt)

	out := make([]float64, len(samples))
	var prev float64
	for i, s := range samples {
		prev = prev + alpha*(s-prev)
		out[i] = prev
	}
	return out
}

// GenerateTone synthesizes a single waveform with ADSR envelope and frequency curve.
func GenerateTone(
	sampleRate int,
	duration float64,
	waveType WaveType,
	dutyCycle float64,
	freqFn func(t float64) float64,
	env ADSREnvelope,
) []float64 {
	numSamples := int(float64(sampleRate) * duration)
	if numSamples <= 0 {
		return nil
	}

	samples := make([]float64, numSamples)
	var phase float64

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		freq := freqFn(t)
		amp := env.ValueAt(t, duration)

		var wave float64
		switch waveType {
		case WaveSine:
			wave = Sine(phase)
		case WaveSquare:
			wave = Square(phase, dutyCycle)
		case WaveSawtooth:
			wave = Sawtooth(phase)
		case WaveTriangle:
			wave = Triangle(phase)
		case WaveNoise:
			wave = Noise()
		}

		samples[i] = wave * amp

		// Advance phase based on instantaneous frequency
		phase += 2.0 * math.Pi * freq / float64(sampleRate)
		if phase >= 2.0*math.Pi {
			phase = math.Mod(phase, 2.0*math.Pi)
		}
	}

	return samples
}

// GenerateFMSweep synthesizes a frequency-modulated tone with carrier and modulator frequency sweeps.
func GenerateFMSweep(
	sampleRate int,
	duration float64,
	carrierFreqFn func(t float64) float64,
	modFreqFn func(t float64) float64,
	modIndexFn func(t float64) float64,
	env ADSREnvelope,
) []float64 {
	numSamples := int(float64(sampleRate) * duration)
	if numSamples <= 0 {
		return nil
	}

	samples := make([]float64, numSamples)
	var carrierPhase float64
	var modPhase float64

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		carrierFreq := carrierFreqFn(t)
		modFreq := modFreqFn(t)
		modIndex := modIndexFn(t)
		amp := env.ValueAt(t, duration)

		// Calculate modulator output
		modVal := math.Sin(modPhase)

		// Modulate carrier frequency
		instFreq := carrierFreq + (modVal * modIndex * modFreq)
		if instFreq < 0 {
			instFreq = 0
		}

		samples[i] = math.Sin(carrierPhase) * amp

		// Advance phases
		carrierPhase += 2.0 * math.Pi * instFreq / float64(sampleRate)
		if carrierPhase >= 2.0*math.Pi {
			carrierPhase = math.Mod(carrierPhase, 2.0*math.Pi)
		}

		modPhase += 2.0 * math.Pi * modFreq / float64(sampleRate)
		if modPhase >= 2.0*math.Pi {
			modPhase = math.Mod(modPhase, 2.0*math.Pi)
		}
	}

	return samples
}

// FloatsToStereoPCM16 converts mono float64 samples [-1.0, 1.0] to 16-bit signed Little-Endian stereo PCM bytes.
func FloatsToStereoPCM16(samples []float64) []byte {
	buf := make([]byte, len(samples)*BytesPerSampleFrame)
	for i, s := range samples {
		// Hard limiter clamp to prevent distortion / wrap-around
		if s > 1.0 {
			s = 1.0
		} else if s < -1.0 {
			s = -1.0
		}
		val := int16(s * 32767.0)

		// Left channel
		buf[i*4] = byte(val)
		buf[i*4+1] = byte(val >> 8)
		// Right channel
		buf[i*4+2] = byte(val)
		buf[i*4+3] = byte(val >> 8)
	}
	return buf
}

// StereoFloatsToStereoPCM16 converts left and right float64 channels to 16-bit Little-Endian stereo PCM bytes.
func StereoFloatsToStereoPCM16(left, right []float64) []byte {
	n := len(left)
	if len(right) < n {
		n = len(right)
	}
	buf := make([]byte, n*BytesPerSampleFrame)
	for i := 0; i < n; i++ {
		l := left[i]
		r := right[i]
		if l > 1.0 {
			l = 1.0
		} else if l < -1.0 {
			l = -1.0
		}
		if r > 1.0 {
			r = 1.0
		} else if r < -1.0 {
			r = -1.0
		}

		valL := int16(l * 32767.0)
		valR := int16(r * 32767.0)

		buf[i*4] = byte(valL)
		buf[i*4+1] = byte(valL >> 8)
		buf[i*4+2] = byte(valR)
		buf[i*4+3] = byte(valR >> 8)
	}
	return buf
}

// VoicePool manages a round-robin pool of ebiten audio players for polyphonic playback.
type VoicePool struct {
	context *ebitenaudio.Context
	players []*ebitenaudio.Player
	head    int
	mu      sync.Mutex
}

// NewVoicePool creates a voice pool with voiceCount players for the specified PCM data.
// If ctx is nil, a silent mock-free pool is returned that safely discards playback calls.
func NewVoicePool(ctx *ebitenaudio.Context, pcmData []byte, voiceCount int) (*VoicePool, error) {
	if voiceCount <= 0 {
		return nil, errors.New("voiceCount must be greater than 0")
	}
	if len(pcmData) == 0 {
		return nil, errors.New("pcmData cannot be empty")
	}

	vp := &VoicePool{
		context: ctx,
	}

	if ctx != nil {
		vp.players = make([]*ebitenaudio.Player, voiceCount)
		for i := 0; i < voiceCount; i++ {
			p := ebitenaudio.NewPlayerFromBytes(ctx, pcmData)
			if p == nil {
				return nil, fmt.Errorf("failed to create audio player %d", i)
			}
			vp.players[i] = p
		}
	}

	return vp, nil
}

// Play triggers an available voice or rewinds the oldest voice.
func (vp *VoicePool) Play() {
	if vp == nil {
		return
	}
	vp.mu.Lock()
	defer vp.mu.Unlock()

	if len(vp.players) == 0 {
		return // Silent / headless mode
	}

	// Try to find a player that is not currently playing
	for i := 0; i < len(vp.players); i++ {
		idx := (vp.head + i) % len(vp.players)
		p := vp.players[idx]
		if !p.IsPlaying() {
			_ = p.Rewind()
			p.Play()
			vp.head = (idx + 1) % len(vp.players)
			return
		}
	}

	// All players are busy: steal the head player and rewind
	p := vp.players[vp.head]
	_ = p.Rewind()
	p.Play()
	vp.head = (vp.head + 1) % len(vp.players)
}

// Close closes all underlying audio players in the pool.
func (vp *VoicePool) Close() error {
	if vp == nil {
		return nil
	}
	vp.mu.Lock()
	defer vp.mu.Unlock()

	var firstErr error
	for _, p := range vp.players {
		if p != nil {
			if err := p.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	vp.players = nil
	return firstErr
}

// GetOrCreateAudioContext safely retrieves the current audio context or initializes a new one.
// Returns an error if audio context creation fails or panics.
func GetOrCreateAudioContext(sampleRate int) (*ebitenaudio.Context, error) {
	if ctx := ebitenaudio.CurrentContext(); ctx != nil {
		return ctx, nil
	}

	var ctx *ebitenaudio.Context
	var initErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				initErr = fmt.Errorf("audio context panic: %v", r)
			}
		}()
		ctx = ebitenaudio.NewContext(sampleRate)
	}()

	if initErr != nil {
		return nil, initErr
	}
	if ctx == nil {
		return nil, errors.New("audio.NewContext returned nil")
	}

	return ctx, nil
}
