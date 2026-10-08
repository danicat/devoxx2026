package entity

import (
	"fmt"
	"testing"
)

func TestNewWaveManager(t *testing.T) {
	wm := NewWaveManager()
	if wm == nil {
		t.Fatal("expected non-nil WaveManager")
	}

	if len(wm.Waves) != 10 {
		t.Fatalf("expected 10 waves, got %d", len(wm.Waves))
	}

	expectedWaves := []struct {
		waveNum int
		name    string
		reward  int
	}{
		{1, "Warmup / Minor Allocation", 100},
		{2, "String Concatenation Storm", 140},
		{3, "Young Gen Pressure", 180},
		{4, "ThreadLocal Leak", 220},
		{5, "Mid-Benchmark Spike", 280},
		{6, "Survivor Space Overflow", 340},
		{7, "Memory Leak Cascade", 400},
		{8, "Native Buffer Allocation", 500},
		{9, "Full GC Stress Test", 650},
		{10, "Catastrophic Allocation Surge", 1000},
	}

	for i, exp := range expectedWaves {
		w := wm.Waves[i]
		if w.WaveNumber != exp.waveNum {
			t.Errorf("wave %d: expected WaveNumber %d, got %d", i+1, exp.waveNum, w.WaveNumber)
		}
		if w.Name != exp.name {
			t.Errorf("wave %d: expected Name %q, got %q", i+1, exp.name, w.Name)
		}
		if w.Reward != exp.reward {
			t.Errorf("wave %d: expected Reward %d, got %d", i+1, exp.reward, w.Reward)
		}
		if len(w.Spawns) == 0 {
			t.Errorf("wave %d: expected non-empty Spawns", i+1)
		}
	}

	if wm.CurrentWaveIndex != -1 {
		t.Errorf("expected initial CurrentWaveIndex -1, got %d", wm.CurrentWaveIndex)
	}
	if wm.Active {
		t.Errorf("expected initially inactive")
	}
	if wm.Completed {
		t.Errorf("expected initially not completed")
	}
	if wm.GetCurrentWaveDef() != nil {
		t.Errorf("expected nil GetCurrentWaveDef before starting")
	}
}

func TestStartWave_Invalid(t *testing.T) {
	wm := NewWaveManager()

	invalidIndices := []int{-1, 0, 11, 99}
	for _, idx := range invalidIndices {
		t.Run(fmt.Sprintf("wave_%d", idx), func(t *testing.T) {
			err := wm.StartWave(idx)
			if err == nil {
				t.Fatalf("expected error for invalid wave number %d, got nil", idx)
			}
		})
	}
}

func TestStartWave_And_UpdateProgression(t *testing.T) {
	wm := NewWaveManager()
	startX, startY := 30.0, 60.0

	err := wm.StartWave(1)
	if err != nil {
		t.Fatalf("unexpected error starting wave 1: %v", err)
	}

	if !wm.IsWaveActive() {
		t.Fatal("expected wave to be active")
	}
	if wm.CurrentWaveIndex != 0 {
		t.Fatalf("expected CurrentWaveIndex 0, got %d", wm.CurrentWaveIndex)
	}
	if wm.GetCurrentWaveDef() == nil || wm.GetCurrentWaveDef().WaveNumber != 1 {
		t.Fatalf("expected CurrentWaveDef wave 1, got %+v", wm.GetCurrentWaveDef())
	}

	// Wave 1 config: 10x ObjLambda, Interval: 0.8s, Delay: 0.5s
	// dt = 0.2s -> total time = 0.2s (before delay 0.5s) -> 0 spawns
	spawned := wm.Update(0.2, startX, startY)
	if len(spawned) != 0 {
		t.Fatalf("expected 0 spawns before delay, got %d", len(spawned))
	}
	if wm.IsWaveSpawningComplete() {
		t.Fatal("wave spawning should not be complete yet")
	}

	// dt = 0.3s -> total time = 0.5s (reached delay 0.5s) -> 1st spawn
	spawned = wm.Update(0.3, startX, startY)
	if len(spawned) != 1 {
		t.Fatalf("expected 1 spawn at delay, got %d", len(spawned))
	}
	first := spawned[0]
	if first.ID != 1 {
		t.Errorf("expected ID 1, got %d", first.ID)
	}
	if first.Type != ObjLambda {
		t.Errorf("expected ObjLambda, got %v", first.Type)
	}
	if first.X != startX || first.Y != startY {
		t.Errorf("expected position (%v, %v), got (%v, %v)", startX, startY, first.X, first.Y)
	}

	// Step forward by 0.8s intervals until all 10 are spawned
	var totalSpawned []*AllocationObject
	totalSpawned = append(totalSpawned, spawned...)

	// We need 9 more spawns, each 0.8s apart
	for step := 0; step < 9; step++ {
		res := wm.Update(0.8, startX, startY)
		if len(res) != 1 {
			t.Fatalf("step %d: expected 1 spawn per 0.8s, got %d", step, len(res))
		}
		totalSpawned = append(totalSpawned, res...)
	}

	if len(totalSpawned) != 10 {
		t.Fatalf("expected 10 total spawned objects, got %d", len(totalSpawned))
	}

	// Check IDs are sequential 1 to 10
	for i, obj := range totalSpawned {
		expectedID := int64(i + 1)
		if obj.ID != expectedID {
			t.Errorf("object %d: expected ID %d, got %d", i, expectedID, obj.ID)
		}
		if obj.Type != ObjLambda {
			t.Errorf("object %d: expected ObjLambda, got %v", i, obj.Type)
		}
	}

	if !wm.IsWaveSpawningComplete() {
		t.Fatal("expected IsWaveSpawningComplete to be true after 10 objects spawned")
	}

	// Further updates should not spawn any more objects
	res := wm.Update(1.0, startX, startY)
	if len(res) != 0 {
		t.Fatalf("expected 0 spawns after complete, got %d", len(res))
	}
}

func TestWaveManager_FullProgression(t *testing.T) {
	wm := NewWaveManager()
	startX, startY := 10.0, 20.0

	for wave := 1; wave <= 10; wave++ {
		err := wm.StartWave(wave)
		if err != nil {
			t.Fatalf("failed to start wave %d: %v", wave, err)
		}

		currentDef := wm.GetCurrentWaveDef()
		expectedTotalCount := 0
		for _, s := range currentDef.Spawns {
			expectedTotalCount += s.Count
		}

		reward := wm.GetWaveReward(wave)
		if reward != currentDef.Reward {
			t.Errorf("wave %d: expected reward %d, got %d", wave, currentDef.Reward, reward)
		}

		// Advance time in small chunks until spawning is complete
		spawnedInWave := 0
		// Max time safety guard (e.g. 60 seconds of wave time)
		for tElapsed := 0.0; tElapsed < 60.0; tElapsed += 0.1 {
			objs := wm.Update(0.1, startX, startY)
			spawnedInWave += len(objs)
			if wm.IsWaveSpawningComplete() {
				break
			}
		}

		if !wm.IsWaveSpawningComplete() {
			t.Fatalf("wave %d did not complete spawning within 60s", wave)
		}
		if spawnedInWave != expectedTotalCount {
			t.Fatalf("wave %d: expected %d spawned objects, got %d", wave, expectedTotalCount, spawnedInWave)
		}

		if wave < 10 {
			if wm.IsAllWavesCompleted() {
				t.Fatalf("wave %d: expected IsAllWavesCompleted to be false", wave)
			}
		}
	}

	if !wm.IsAllWavesCompleted() {
		t.Fatalf("expected IsAllWavesCompleted to be true after wave 10 finished")
	}

	// Trying to start wave 11 via StartWave should fail
	if err := wm.StartWave(11); err == nil {
		t.Fatal("expected error starting wave 11")
	}
}

func TestStartNextWave(t *testing.T) {
	wm := NewWaveManager()

	// Initial start next wave should start wave 1
	err := wm.StartNextWave()
	if err != nil {
		t.Fatalf("unexpected error on first StartNextWave: %v", err)
	}
	if wm.CurrentWaveIndex != 0 {
		t.Fatalf("expected CurrentWaveIndex 0, got %d", wm.CurrentWaveIndex)
	}

	// Advance through all 10 waves
	for w := 2; w <= 10; w++ {
		err := wm.StartNextWave()
		if err != nil {
			t.Fatalf("expected to start wave %d, got err: %v", w, err)
		}
		if wm.CurrentWaveIndex != w-1 {
			t.Fatalf("expected CurrentWaveIndex %d, got %d", w-1, wm.CurrentWaveIndex)
		}
	}

	// Next wave beyond 10 should return fail-fast error
	err = wm.StartNextWave()
	if err == nil {
		t.Fatal("expected error starting wave beyond 10, got nil")
	}
	if !wm.Completed {
		t.Fatal("expected wm.Completed to be true after exhausted waves")
	}
}

func TestGetWaveReward_Invalid(t *testing.T) {
	wm := NewWaveManager()
	if wm.GetWaveReward(0) != 0 {
		t.Errorf("expected 0 reward for wave 0")
	}
	if wm.GetWaveReward(11) != 0 {
		t.Errorf("expected 0 reward for wave 11")
	}
}
