# JVM GC: Heap Defender - Architecture & Interface Specification

This document defines the system architecture, domain boundaries, and Go package contracts for **JVM GC: Heap Defender**. All domain leads and specialist workers must adhere strictly to these contracts.

---

## 1. Directory & Package Structure

```text
05-skills/
├── cmd/
│   └── game/
│       └── main.go                  # Main entry point (ebiten.RunGameWithOptions)
├── internal/
│   ├── heap/                        # Heap memory model, lanes, grid coordinates
│   │   ├── grid.go
│   │   ├── lane.go
│   │   └── waypoint.go
│   ├── entity/                      # Objects (enemies) & Collectors (towers)
│   │   ├── object.go                # Memory allocations (Lambda, String, ThreadLocal, LargeBlob)
│   │   ├── collector.go             # GC towers (Serial, Parallel, CMS, G1, ZGC)
│   │   ├── projectile.go            # Laser rays & particle beams
│   │   └── wave.go                  # Wave definitions & spawner logic
│   ├── art/                         # Pure-code procedural graphics & shaders
│   │   ├── textures.go              # Procedurally baked bitmap textures
│   │   ├── vector.go                # Real-time vector drawing helpers
│   │   └── particles.go             # Additive particle pool & emitter
│   ├── audio/                       # Pure-code DSP synthesizer
│   │   ├── synth.go                 # Wave generators (sine, square, saw, noise)
│   │   └── sfx.go                   # Procedural SFX triggers (sweep, blip, pop, alarm)
│   ├── ui/                          # HUD, 9-slice panels, buttons, telemetry gauges
│   │   ├── hud.go                   # Telemetry bars (Eden, Survivor, Old Gen, CPU)
│   │   ├── button.go                # Interactive buttons & hotkeys
│   │   └── inspector.go             # Tower inspector & upgrade panel
│   └── game/                        # Core ebiten.Game, state machine, scene manager
│       ├── game.go                  # ebiten.Game implementation
│       ├── state.go                 # Scene FSM interface & transitions
│       └── play_scene.go            # Main play scene coordinating subsystems
└── docs/
    └── ARCHITECTURE.md
```

---

## 2. Global Game Constants & Configuration

- **Virtual Canvas**: `Width = 960`, `Height = 540` (16:9 widescreen)
- **Target TPS**: `60` (`ebiten.SetTPS(60)`)
- **Grid Layout**:
  - Tile size: `30x30` pixels (`32x18` grid)
  - Playing Arena: `X: 30` to `X: 750`, `Y: 45` to `Y: 495`
  - HUD / Dock Area: `X: 760` to `X: 950` (Right Inspector) and Top Header `Y: 0` to `Y: 45`

---

## 3. Data Contracts & Interfaces

### 3.1 Heap & Memory Generation Lanes (`internal/heap`)

```go
package heap

type GenerationType int

const (
    GenEden GenerationType = iota
    GenSurvivor0
    GenSurvivor1
    GenTenured
)

type Waypoint struct {
    X, Y float64
}

type MemoryPath struct {
    Waypoints []Waypoint
    Length    float64
}

type LaneConfig struct {
    GenType    GenerationType
    ColorHex   uint32
    CapacityMB int
}
```

### 3.2 Entities: Allocations & Collectors (`internal/entity`)

```go
package entity

import "github.com/hajimehoshi/ebiten/v2"

type ObjectType int

const (
    ObjLambda ObjectType = iota  // Fast, low HP, ephemeral
    ObjString                    // Medium HP, prone to promotion
    ObjThreadLocal               // Armored, cyclic references
    ObjLargeBlob                 // Boss heavyweight allocation
)

type AllocationObject struct {
    ID          int64
    Type        ObjectType
    SizeMB      int
    MaxHealth   float64
    Health      float64
    Speed       float64
    PathIndex   int
    X, Y        float64
    TenureAge   int
    IsDead      bool
    ReachedEnd  bool
}

type CollectorType int

const (
    ColSerial CollectorType = iota   // Single-target beam
    ColParallel                      // High-throughput burst
    ColCMS                           // Radial mark-sweep
    ColG1                            // Prioritizes densest garbage
    ColZGC                           // High-speed low-latency laser
)

type Collector struct {
    ID         int64
    Type       CollectorType
    GridX, GridY int
    WorldX, WorldY float64
    Range      float64
    DPS        float64
    FireRate   float64
    Cooldown   float64
    TargetID   int64
    Level      int
    Cost       int
}
```

### 3.3 Audio Synthesizer Contracts (`internal/audio`)

```go
package audio

type AudioManager interface {
    PlaySFXSweep()
    PlaySFXAllocate()
    PlaySFXDeallocate()
    PlaySFXPromote()
    PlaySFXSTW()
    PlaySFXOOMAlarm()
    Update()
}
```

### 3.4 Procedural Art Contracts (`internal/art`)

```go
package art

import "github.com/hajimehoshi/ebiten/v2"

type ProceduralAssets struct {
    SiliconTile     *ebiten.Image
    BusTrace        *ebiten.Image
    CollectorTurret map[int]*ebiten.Image
    ObjectSprites   map[int]*ebiten.Image
}

func GenerateAssets() (*ProceduralAssets, error)
```

---

## 4. Swarm Engineering Guardrails

1. **Strict Fail-Fast**: Never mask errors or implement fallback stubs. If initialization fails, return descriptive errors immediately.
2. **Zero File Clashes**: Domains own strictly disjoint packages as allocated above.
3. **No External Assets**: No `.png`, `.jpg`, `.mp3`, or `.wav` files. 100% pure-code procedural rendering and audio synthesis.
4. **WASM Compatibility**: Standard Ebitengine v2 APIs compatible with `GOOS=js GOARCH=wasm`.
