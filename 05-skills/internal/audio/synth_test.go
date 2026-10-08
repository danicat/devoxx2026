package audio

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestSineWave(t *testing.T) {
	epsilon := 1e-6
	cases := []struct {
		phase    float64
		expected float64
	}{
		{0, 0.0},
		{math.Pi / 2, 1.0},
		{math.Pi, 0.0},
		{3 * math.Pi / 2, -1.0},
		{2 * math.Pi, 0.0},
	}

	for _, c := range cases {
		got := Sine(c.phase)
		if math.Abs(got-c.expected) > epsilon {
			t.Errorf("Sine(%f) = %f, expected %f", c.phase, got, c.expected)
		}
	}
}

func TestSquareWave(t *testing.T) {
	// 50% duty cycle
	if got := Square(0.1, 0.5); got != 1.0 {
		t.Errorf("Square(0.1, 0.5) = %f, expected 1.0", got)
	}
	if got := Square(math.Pi+0.1, 0.5); got != -1.0 {
		t.Errorf("Square(pi+0.1, 0.5) = %f, expected -1.0", got)
	}

	// 25% duty cycle
	// phase corresponding to 20% of cycle (norm = 0.20 < 0.25) -> 1.0
	phase20 := 0.20 * 2.0 * math.Pi
	if got := Square(phase20, 0.25); got != 1.0 {
		t.Errorf("Square(phase20, 0.25) = %f, expected 1.0", got)
	}
	// phase corresponding to 30% of cycle (norm = 0.30 >= 0.25) -> -1.0
	phase30 := 0.30 * 2.0 * math.Pi
	if got := Square(phase30, 0.25); got != -1.0 {
		t.Errorf("Square(phase30, 0.25) = %f, expected -1.0", got)
	}

	// Invalid duty cycle defaults to 0.5
	if got := Square(0.1, -0.5); got != 1.0 {
		t.Errorf("Square(0.1, -0.5) with invalid duty cycle expected 1.0, got %f", got)
	}
}

func TestSawtoothWave(t *testing.T) {
	epsilon := 1e-4

	// At start of cycle (norm ~0.0): close to -1.0
	if got := Sawtooth(0.001); math.Abs(got-(-1.0)) > 0.01 {
		t.Errorf("Sawtooth at start = %f, expected near -1.0", got)
	}

	// At mid cycle (norm = 0.5, phase = Pi): should be 0.0
	if got := Sawtooth(math.Pi); math.Abs(got-0.0) > epsilon {
		t.Errorf("Sawtooth(Pi) = %f, expected 0.0", got)
	}

	// Range bounds across full cycle
	for step := 0; step < 100; step++ {
		p := float64(step) / 100.0 * 2.0 * math.Pi
		val := Sawtooth(p)
		if val < -1.0 || val > 1.0 {
			t.Errorf("Sawtooth out of bounds: %f at phase %f", val, p)
		}
	}
}

func TestTriangleWave(t *testing.T) {
	epsilon := 1e-5

	// Phase 0 -> 1.0
	if got := Triangle(0); math.Abs(got-1.0) > epsilon {
		t.Errorf("Triangle(0) = %f, expected 1.0", got)
	}
	// Phase Pi/2 (norm 0.25) -> 0.0
	if got := Triangle(math.Pi / 2); math.Abs(got-0.0) > epsilon {
		t.Errorf("Triangle(Pi/2) = %f, expected 0.0", got)
	}
	// Phase Pi (norm 0.5) -> -1.0
	if got := Triangle(math.Pi); math.Abs(got-(-1.0)) > epsilon {
		t.Errorf("Triangle(Pi) = %f, expected -1.0", got)
	}
	// Phase 3Pi/2 (norm 0.75) -> 0.0
	if got := Triangle(3 * math.Pi / 2); math.Abs(got-0.0) > epsilon {
		t.Errorf("Triangle(3Pi/2) = %f, expected 0.0", got)
	}
}

func TestNoiseGenerators(t *testing.T) {
	// Test global Noise generator bounds
	for i := 0; i < 500; i++ {
		val := Noise()
		if val < -1.0 || val > 1.0 {
			t.Fatalf("Noise() value out of bounds [-1.0, 1.0]: %f", val)
		}
	}

	// Test DeterministicNoise reproducibility
	state1 := uint64(123456789)
	state2 := uint64(123456789)

	for i := 0; i < 100; i++ {
		v1 := DeterministicNoise(&state1)
		v2 := DeterministicNoise(&state2)
		if v1 != v2 {
			t.Fatalf("DeterministicNoise non-deterministic at step %d: %f != %f", i, v1, v2)
		}
	}
}

func TestADSREnvelope(t *testing.T) {
	env := ADSREnvelope{
		AttackDuration:  0.02,
		DecayDuration:   0.03,
		SustainLevel:    0.4,
		ReleaseDuration: 0.05,
	}
	total := 0.20 // 0.00-0.02 attack, 0.02-0.05 decay, 0.05-0.15 sustain, 0.15-0.20 release

	// Boundary conditions
	if got := env.ValueAt(-0.01, total); got != 0.0 {
		t.Errorf("ValueAt before start = %f, expected 0.0", got)
	}
	if got := env.ValueAt(total, total); got != 0.0 {
		t.Errorf("ValueAt at/past end = %f, expected 0.0", got)
	}
	if got := env.ValueAt(0.0, total); got != 0.0 {
		t.Errorf("ValueAt at start = %f, expected 0.0", got)
	}

	// End of Attack (t = 0.02)
	gotAttackPeak := env.ValueAt(0.02, total)
	if math.Abs(gotAttackPeak-1.0) > 0.01 {
		t.Errorf("ValueAt attack peak = %f, expected near 1.0", gotAttackPeak)
	}

	// End of Decay / Start of Sustain (t = 0.05)
	gotSustain := env.ValueAt(0.05, total)
	if math.Abs(gotSustain-0.4) > 0.01 {
		t.Errorf("ValueAt sustain start = %f, expected near 0.4", gotSustain)
	}

	// Mid sustain (t = 0.10)
	gotMidSustain := env.ValueAt(0.10, total)
	if math.Abs(gotMidSustain-0.4) > 0.01 {
		t.Errorf("ValueAt mid sustain = %f, expected 0.4", gotMidSustain)
	}

	// Mid release (t = 0.175, halfway through 0.15..0.20 release, so 0.4 * 0.5 = 0.2)
	gotMidRelease := env.ValueAt(0.175, total)
	if math.Abs(gotMidRelease-0.2) > 0.02 {
		t.Errorf("ValueAt mid release = %f, expected near 0.2", gotMidRelease)
	}
}

func TestADSREnvelopeOvershootScaling(t *testing.T) {
	// Total duration is smaller than sum of phases (0.05 total < 0.10 attack + 0.10 decay)
	env := ADSREnvelope{
		AttackDuration:  0.10,
		DecayDuration:   0.10,
		SustainLevel:    0.5,
		ReleaseDuration: 0.10,
	}
	// Should not panic, should scale cleanly and stay bounded in [0, 1]
	for i := 0; i < 50; i++ {
		tSec := float64(i) / 1000.0
		val := env.ValueAt(tSec, 0.05)
		if val < 0.0 || val > 1.0 {
			t.Errorf("ValueAt out of [0, 1] bounds during scaled envelope: %f", val)
		}
	}
}

func TestApplyLowPassFilter(t *testing.T) {
	// Empty slice
	if filtered := ApplyLowPassFilter(nil, 44100, 1000); filtered != nil {
		t.Errorf("Expected nil for nil input")
	}

	// High frequency signal (10kHz) filtered with 200Hz cutoff should be significantly attenuated
	numSamples := 4410
	highFreq := make([]float64, numSamples)
	for i := range highFreq {
		highFreq[i] = math.Sin(2.0 * math.Pi * 10000.0 * float64(i) / 44100.0)
	}

	filtered := ApplyLowPassFilter(highFreq, 44100, 200.0)
	if len(filtered) != numSamples {
		t.Fatalf("Filtered sample length mismatch: %d != %d", len(filtered), numSamples)
	}

	// Measure amplitude after filter settles
	var maxAmp float64
	for i := numSamples / 2; i < numSamples; i++ {
		if abs := math.Abs(filtered[i]); abs > maxAmp {
			maxAmp = abs
		}
	}

	if maxAmp > 0.15 {
		t.Errorf("10kHz signal was not sufficiently attenuated by 200Hz LPF, maxAmp: %f", maxAmp)
	}
}

func TestFloatsToStereoPCM16(t *testing.T) {
	samples := []float64{0.0, 1.0, -1.0, 1.5, -2.0}
	pcm := FloatsToStereoPCM16(samples)

	expectedLen := len(samples) * BytesPerSampleFrame
	if len(pcm) != expectedLen {
		t.Fatalf("PCM length %d != expected %d", len(pcm), expectedLen)
	}

	// Sample 0: 0.0 -> int16(0)
	val0L := int16(binary.LittleEndian.Uint16(pcm[0:2]))
	val0R := int16(binary.LittleEndian.Uint16(pcm[2:4]))
	if val0L != 0 || val0R != 0 {
		t.Errorf("Sample 0 expected 0, got L=%d R=%d", val0L, val0R)
	}

	// Sample 1: 1.0 -> 32767
	val1L := int16(binary.LittleEndian.Uint16(pcm[4:6]))
	if val1L != 32767 {
		t.Errorf("Sample 1 expected 32767, got %d", val1L)
	}

	// Sample 2: -1.0 -> -32767
	val2L := int16(binary.LittleEndian.Uint16(pcm[8:10]))
	if val2L != -32767 {
		t.Errorf("Sample 2 expected -32767, got %d", val2L)
	}

	// Sample 3: 1.5 -> clamped to 32767
	val3L := int16(binary.LittleEndian.Uint16(pcm[12:14]))
	if val3L != 32767 {
		t.Errorf("Sample 3 clamped expected 32767, got %d", val3L)
	}

	// Sample 4: -2.0 -> clamped to -32767
	val4L := int16(binary.LittleEndian.Uint16(pcm[16:18]))
	if val4L != -32767 {
		t.Errorf("Sample 4 clamped expected -32767, got %d", val4L)
	}
}

func TestStereoFloatsToStereoPCM16(t *testing.T) {
	left := []float64{0.5, -0.5}
	right := []float64{-0.25, 0.75}

	pcm := StereoFloatsToStereoPCM16(left, right)
	if len(pcm) != 2*BytesPerSampleFrame {
		t.Fatalf("Expected %d bytes, got %d", 2*BytesPerSampleFrame, len(pcm))
	}

	l0 := int16(binary.LittleEndian.Uint16(pcm[0:2]))
	r0 := int16(binary.LittleEndian.Uint16(pcm[2:4]))

	expectedL0 := int16(math.Round(0.5 * 32767.0))
	expectedR0 := int16(math.Round(-0.25 * 32767.0))

	if math.Abs(float64(l0-expectedL0)) > 2 {
		t.Errorf("Stereo left channel mismatch: got %d, expected %d", l0, expectedL0)
	}
	if math.Abs(float64(r0-expectedR0)) > 2 {
		t.Errorf("Stereo right channel mismatch: got %d, expected %d", r0, expectedR0)
	}
}

func TestGenerateToneAndFMSweep(t *testing.T) {
	env := ADSREnvelope{
		AttackDuration:  0.01,
		DecayDuration:   0.02,
		SustainLevel:    0.5,
		ReleaseDuration: 0.02,
	}

	// Test GenerateTone
	tone := GenerateTone(44100, 0.1, WaveSine, 0.5, func(t float64) float64 { return 440.0 }, env)
	if len(tone) != 4410 {
		t.Fatalf("Tone sample count = %d, expected 4410", len(tone))
	}
	for i, s := range tone {
		if s < -1.0 || s > 1.0 {
			t.Errorf("Tone sample out of bounds at %d: %f", i, s)
		}
	}

	// Test GenerateFMSweep
	fm := GenerateFMSweep(
		44100,
		0.05,
		func(t float64) float64 { return 800.0 - 400.0*t },
		func(t float64) float64 { return 150.0 },
		func(t float64) float64 { return 2.0 },
		env,
	)
	if len(fm) != 2205 {
		t.Fatalf("FM sweep sample count = %d, expected 2205", len(fm))
	}
	for i, s := range fm {
		if s < -1.0 || s > 1.0 {
			t.Errorf("FM sample out of bounds at %d: %f", i, s)
		}
	}
}

func TestVoicePoolHeadless(t *testing.T) {
	// Negative voice count should return error
	_, err := NewVoicePool(nil, []byte{1, 2, 3, 4}, 0)
	if err == nil {
		t.Errorf("Expected error for voiceCount <= 0")
	}

	// Empty PCM should return error
	_, err = NewVoicePool(nil, nil, 3)
	if err == nil {
		t.Errorf("Expected error for empty PCM")
	}

	// Headless pool (nil context)
	pcm := make([]byte, 16)
	vp, err := NewVoicePool(nil, pcm, 3)
	if err != nil {
		t.Fatalf("Failed to create headless voice pool: %v", err)
	}

	// Play and Close should not panic in headless mode
	vp.Play()
	vp.Play()
	if err := vp.Close(); err != nil {
		t.Errorf("Unexpected error closing headless voice pool: %v", err)
	}
}
