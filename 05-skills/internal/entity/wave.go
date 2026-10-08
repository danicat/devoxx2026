package entity

import (
	"fmt"
)

// SpawnEntry defines a single spawning schedule for a specific ObjectType within a wave.
type SpawnEntry struct {
	Type     ObjectType
	Count    int
	Interval float64 // Delay between each spawn of this entry in seconds
	Delay    float64 // Initial delay before this entry starts spawning in seconds
}

// WaveDefinition specifies the configuration, enemy profile, and reward of a wave.
type WaveDefinition struct {
	WaveNumber  int
	Name        string
	Description string
	Spawns      []SpawnEntry
	Reward      int // CPU Cycles awarded upon clearing wave
}

// WaveManager manages the progression of waves, tracking spawn timers and generating AllocationObjects.
type WaveManager struct {
	Waves            []WaveDefinition
	CurrentWaveIndex int
	Active           bool
	WaveTime         float64
	SpawnTimers      []float64
	SpawnedCounts    []int
	NextObjectID     int64
	Completed        bool
}

// DefaultWaveDefinitions returns the standard 10 structured allocation waves.
func DefaultWaveDefinitions() []WaveDefinition {
	return []WaveDefinition{
		{
			WaveNumber:  1,
			Name:        "Warmup / Minor Allocation",
			Description: "High-frequency ephemeral short-lived Lambda objects flooding Eden space.",
			Reward:      100,
			Spawns: []SpawnEntry{
				{Type: ObjLambda, Count: 10, Interval: 0.8, Delay: 0.5},
			},
		},
		{
			WaveNumber:  2,
			Name:        "String Concatenation Storm",
			Description: "Unbuffered string builders creating medium-lived String allocations mixed with Lambdas.",
			Reward:      140,
			Spawns: []SpawnEntry{
				{Type: ObjLambda, Count: 8, Interval: 0.7, Delay: 0.3},
				{Type: ObjString, Count: 6, Interval: 1.2, Delay: 1.0},
			},
		},
		{
			WaveNumber:  3,
			Name:        "Young Gen Pressure",
			Description: "Intense allocation burst threatening to overflow Eden space into Survivor spaces.",
			Reward:      180,
			Spawns: []SpawnEntry{
				{Type: ObjString, Count: 12, Interval: 0.9, Delay: 0.5},
				{Type: ObjLambda, Count: 10, Interval: 0.6, Delay: 0.8},
			},
		},
		{
			WaveNumber:  4,
			Name:        "ThreadLocal Leak",
			Description: "Armored cyclic references and ThreadLocal variables resistant to standard sweeping.",
			Reward:      220,
			Spawns: []SpawnEntry{
				{Type: ObjThreadLocal, Count: 6, Interval: 1.5, Delay: 0.5},
				{Type: ObjString, Count: 8, Interval: 1.0, Delay: 1.2},
			},
		},
		{
			WaveNumber:  5,
			Name:        "Mid-Benchmark Spike",
			Description: "Concurrent microservice benchmark spike mixing all standard Young Gen allocation profiles.",
			Reward:      280,
			Spawns: []SpawnEntry{
				{Type: ObjLambda, Count: 10, Interval: 0.6, Delay: 0.2},
				{Type: ObjString, Count: 8, Interval: 0.9, Delay: 0.8},
				{Type: ObjThreadLocal, Count: 6, Interval: 1.4, Delay: 1.5},
			},
		},
		{
			WaveNumber:  6,
			Name:        "Survivor Space Overflow",
			Description: "Aging String buffers and ThreadLocal references causing massive survivor boundary pressure.",
			Reward:      340,
			Spawns: []SpawnEntry{
				{Type: ObjString, Count: 12, Interval: 0.8, Delay: 0.4},
				{Type: ObjThreadLocal, Count: 10, Interval: 1.2, Delay: 1.0},
			},
		},
		{
			WaveNumber:  7,
			Name:        "Memory Leak Cascade",
			Description: "Heavy ThreadLocal collections with unclosed thread pools creating severe tenured risks.",
			Reward:      400,
			Spawns: []SpawnEntry{
				{Type: ObjThreadLocal, Count: 15, Interval: 1.0, Delay: 0.5},
				{Type: ObjLambda, Count: 15, Interval: 0.5, Delay: 0.5},
			},
		},
		{
			WaveNumber:  8,
			Name:        "Native Buffer Allocation",
			Description: "Direct byte buffer boss allocation requiring immediate concentrated GC focus.",
			Reward:      500,
			Spawns: []SpawnEntry{
				{Type: ObjLargeBlob, Count: 1, Interval: 0.0, Delay: 0.5},
				{Type: ObjLambda, Count: 12, Interval: 0.6, Delay: 1.0},
				{Type: ObjString, Count: 8, Interval: 1.0, Delay: 1.5},
			},
		},
		{
			WaveNumber:  9,
			Name:        "Full GC Stress Test",
			Description: "Dual LargeBlob direct allocations surrounded by armored cyclic ThreadLocal graphs.",
			Reward:      650,
			Spawns: []SpawnEntry{
				{Type: ObjLargeBlob, Count: 2, Interval: 5.0, Delay: 0.5},
				{Type: ObjThreadLocal, Count: 12, Interval: 1.0, Delay: 1.0},
				{Type: ObjString, Count: 10, Interval: 0.8, Delay: 1.5},
			},
		},
		{
			WaveNumber:  10,
			Name:        "Catastrophic Allocation Surge",
			Description: "Catastrophic OOM threshold surge: Triple LargeBlob direct buffers and relentless allocation storms.",
			Reward:      1000,
			Spawns: []SpawnEntry{
				{Type: ObjLargeBlob, Count: 3, Interval: 4.0, Delay: 0.5},
				{Type: ObjThreadLocal, Count: 15, Interval: 0.9, Delay: 1.0},
				{Type: ObjString, Count: 20, Interval: 0.6, Delay: 1.2},
				{Type: ObjLambda, Count: 20, Interval: 0.4, Delay: 0.5},
			},
		},
	}
}

// NewWaveManager creates and initializes a WaveManager with all 10 standard waves.
func NewWaveManager() *WaveManager {
	return &WaveManager{
		Waves:            DefaultWaveDefinitions(),
		CurrentWaveIndex: -1,
		Active:           false,
		WaveTime:         0,
		SpawnTimers:      nil,
		SpawnedCounts:    nil,
		NextObjectID:     1,
		Completed:        false,
	}
}

// StartWave begins the wave corresponding to waveNum (1-indexed).
// It resets wave timers and initializes spawn tracking.
func (wm *WaveManager) StartWave(waveNum int) error {
	if waveNum < 1 || waveNum > len(wm.Waves) {
		return fmt.Errorf("invalid wave number %d: must be between 1 and %d", waveNum, len(wm.Waves))
	}

	idx := waveNum - 1
	wm.CurrentWaveIndex = idx
	wm.Active = true
	wm.WaveTime = 0
	wm.Completed = false

	waveDef := wm.Waves[idx]
	spawnsCount := len(waveDef.Spawns)
	wm.SpawnTimers = make([]float64, spawnsCount)
	wm.SpawnedCounts = make([]int, spawnsCount)

	return nil
}

// StartNextWave transitions to and starts the next wave in sequence.
func (wm *WaveManager) StartNextWave() error {
	nextWaveNum := wm.CurrentWaveIndex + 2 // CurrentWaveIndex is 0-indexed (-1 initially)
	if nextWaveNum > len(wm.Waves) {
		wm.Completed = true
		wm.Active = false
		return fmt.Errorf("no more waves: reached end of configured waves (%d)", len(wm.Waves))
	}
	return wm.StartWave(nextWaveNum)
}

// Update advances wave time and checks all spawn entries, returning any newly spawned AllocationObjects.
func (wm *WaveManager) Update(dt float64, startX, startY float64) []*AllocationObject {
	if !wm.Active || wm.CurrentWaveIndex < 0 || wm.CurrentWaveIndex >= len(wm.Waves) {
		return nil
	}

	wm.WaveTime += dt
	waveDef := &wm.Waves[wm.CurrentWaveIndex]
	var spawned []*AllocationObject

	for i := range waveDef.Spawns {
		entry := &waveDef.Spawns[i]
		if wm.SpawnedCounts[i] >= entry.Count {
			continue
		}

		// Check initial delay
		if wm.WaveTime < entry.Delay {
			continue
		}

		if wm.SpawnedCounts[i] == 0 {
			// First spawn ready upon reaching or passing Delay
			obj := NewAllocationObject(wm.NextObjectID, entry.Type, startX, startY)
			wm.NextObjectID++
			spawned = append(spawned, obj)
			wm.SpawnedCounts[i]++
			wm.SpawnTimers[i] = 0 // reset interval timer after spawn
		} else {
			// Subsequent spawns depend on Interval
			wm.SpawnTimers[i] += dt
			interval := entry.Interval
			if interval <= 0 {
				interval = 0.0001 // prevent infinite loop if interval is non-positive
			}
			for wm.SpawnTimers[i] >= interval && wm.SpawnedCounts[i] < entry.Count {
				wm.SpawnTimers[i] -= interval
				obj := NewAllocationObject(wm.NextObjectID, entry.Type, startX, startY)
				wm.NextObjectID++
				spawned = append(spawned, obj)
				wm.SpawnedCounts[i]++
			}
		}
	}

	return spawned
}

// IsWaveSpawningComplete returns true if all scheduled objects for the current active wave have spawned.
func (wm *WaveManager) IsWaveSpawningComplete() bool {
	if wm.CurrentWaveIndex < 0 || wm.CurrentWaveIndex >= len(wm.Waves) {
		return false
	}
	waveDef := &wm.Waves[wm.CurrentWaveIndex]
	if len(wm.SpawnedCounts) < len(waveDef.Spawns) {
		return false
	}
	for i, entry := range waveDef.Spawns {
		if wm.SpawnedCounts[i] < entry.Count {
			return false
		}
	}
	return true
}

// IsWaveActive returns whether a wave is currently actively running.
func (wm *WaveManager) IsWaveActive() bool {
	return wm.Active
}

// IsAllWavesCompleted returns true if the final wave has finished spawning and the manager is completed.
func (wm *WaveManager) IsAllWavesCompleted() bool {
	return wm.Completed || (wm.CurrentWaveIndex == len(wm.Waves)-1 && wm.IsWaveSpawningComplete())
}

// GetCurrentWaveDef returns the WaveDefinition for the currently active wave, or nil if no wave is active.
func (wm *WaveManager) GetCurrentWaveDef() *WaveDefinition {
	if wm.CurrentWaveIndex < 0 || wm.CurrentWaveIndex >= len(wm.Waves) {
		return nil
	}
	return &wm.Waves[wm.CurrentWaveIndex]
}

// GetWaveReward returns the CPU cycle reward for a given 1-indexed wave number. Returns 0 if invalid.
func (wm *WaveManager) GetWaveReward(waveNum int) int {
	if waveNum < 1 || waveNum > len(wm.Waves) {
		return 0
	}
	return wm.Waves[waveNum-1].Reward
}
