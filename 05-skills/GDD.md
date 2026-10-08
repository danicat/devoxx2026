# Game Design Document (GDD)

> **Game Title**: JVM GC: Heap Defender
> **Target Genre**: 2D Strategy / Tower Defense
> **Target Platform**: Desktop (macOS/Linux/Windows) & WebAssembly
> **Target Aspect Ratio & Resolution**: 16:9 Widescreen (`960x540` Virtual Canvas)
> **Author / Lead Designer**: Antigravity Game Designer & Swarm Coordinator

---

## 1. Executive Summary & Elevator Pitch
- **Elevator Pitch**: Defend the Java Virtual Machine from catastrophic `java.lang.OutOfMemoryError` crashes! Deploy and tune Garbage Collectors (Serial, Parallel, CMS, G1, ZGC) across memory spaces (Eden, Survivor S0/S1, Tenured Old Gen) to sweep, mark, and deallocate rogue object allocations before heap capacity blows out.
- **Core Inspiration**: Classic Tower Defense (*Kingdom Rush*, *Bloons TD*) meets low-level JVM memory management internals.
- **Target Audience / Mood**: Fast-paced, tactical strategy celebrating software engineering, runtime internals, and retro-futuristic silicon aesthetics.

---

## 2. Core Gameplay Loop & Mechanics
- **Primary Gameplay Loop**:
  1. **Allocate**: Application threads spawn waves of memory allocations (objects) entering the **Eden Space** lane.
  2. **Sweep & Mark**: Strategically placed GC Towers fire collector beams to reclaim unreferenced memory, earning **CPU Cycles** (currency).
  3. **Promote**: Objects that survive GC cycles increment their tenure age; when threshold is reached, they promote down memory buses into **Survivor Spaces (S0/S1)** and ultimately **Tenured / Old Gen**.
  4. **Compact & Tune**: Spend earned CPU Cycles to upgrade collector thread pools, reduce Stop-The-World (STW) latency, and install advanced regional collectors (G1, ZGC).
- **Core Mechanics**:
  - **Memory Pools as Lanes**: Continuous path traversing Eden -> Survivor S0/S1 -> Tenured Old Gen.
  - **Tenure Age Counter**: Surviving objects visually show their age counter; older objects require heavier collection passes.
  - **Stop-The-World (STW) Super-Ability**: Tactical screen-wide pause button that halts object advancement for 3 seconds while collectors execute a full major compaction.
- **Enemy Object Profiles**:
  - **Ephemeral Lambda / DTO**: High speed, low byte size (low HP), quickly cleared by lightweight collectors.
  - **String / Collection**: Moderate speed, medium HP, easily promoted to Survivor space if not cleared in Eden.
  - **ThreadLocal / Cyclic Graph**: Slow speed, high HP, armored against simple collectors; requires Mark-Sweep or Concurrent scanning.
  - **Giant Native ByteBuffer / Finalizer**: Massive boss unit, high byte size; rapidly fills Old Gen if it reaches the end of the pipeline.
- **Collector Towers**:
  - **Serial GC**: Single-target direct deallocation beam. Cheap, low CPU footprint, ideal early-game collector.
  - **Parallel Scavenger**: High-throughput multi-threaded burst deallocator in Young Gen.
  - **CMS (Concurrent Mark-Sweep)**: Radial pulse collector that marks and sweeps multiple concurrent objects without blocking.
  - **G1GC (Garbage-First)**: Divides memory into regions; automatically locks onto the densest cluster of high-garbage objects.
  - **ZGC / Shenandoah**: Ultra-low-latency colored-pointer laser; cuts through tenured memory leaks with near-zero pause time.
- **Win & Loss Conditions**:
  - **Win Condition**: Clean all 10 allocation benchmark waves, keeping the Old Gen heap intact.
  - **Loss Condition**: Tenured Old Gen capacity reaches 100% -> fatal `OutOfMemoryError` crash.

---

## 3. Controls & Input Mapping Scheme
- **Primary Input Devices**: Mouse & Keyboard (Full Touch support for WASM).
- **Default Control Scheme**:
  | Logical Action | Mouse / Touch | Keyboard Shortcut |
  | :--- | :--- | :--- |
  | `Select Tower Type` | Click HUD icon | `1` (Serial), `2` (Parallel), `3` (CMS), `4` (G1), `5` (ZGC) |
  | `Place Tower` | Left-Click valid grid slot | Hover + `Enter` |
  | `Inspect / Upgrade` | Click existing tower | Click existing tower |
  | `Trigger Full GC (STW)` | Click STW button | `Space` / `F` |
  | `Start Next Wave` | Click "Allocate Wave" | `Space` (between waves) |
  | `Deselect / Cancel` | Right-Click / Tap away | `Escape` |
  | `Fullscreen Toggle` | Click Fullscreen icon | `F11` or `Alt+Enter` |

---

## 4. Visual Style & Asset Strategy
- **Aesthetic Direction**: Glowing Vector Silicon Die & Memory Bus. Neon memory addresses, luminescent phosphor trace lines, color-coded generation boundaries.
  - **Eden Space**: Neon Emerald Green (`#00E676`)
  - **Survivor Spaces (S0/S1)**: Warm Amber / Gold (`#FFD600`)
  - **Tenured Old Gen**: High-alert Crimson / Neon Red (`#FF1744`)
- **Asset Creation Approach (`procedural-art`)**:
  - **Zero external image files**: 100% pure-code procedural rendering using Go `ebiten/vector` and procedural bitmap texture buffers.
  - **Laser Beams & Particle FX**: Additive blended particle bursts on deallocation, glowing sweep rays, and matrix memory trace lines.
  - **Custom Typography**: Procedurally rendered vector digital glyphs and clean readable metrics.

---

## 5. Audio & Soundscape Strategy (`procedural-composer`)
- **Sound Engine (`internal/audio`)**: Zero external MP3/WAV files. Pure-code real-time DSP chiptune and additive synthesizer.
- **Procedural Sound Effects (SFX)**:
  - `SFX_Beam_Sweep`: FM synthesis sweep with slight resonant decay.
  - `SFX_Allocate_Blip`: High-pitch square-wave bleeps on new object spawn.
  - `SFX_Deallocate_Pop`: Noise burst filtered into soft click/pop on object death.
  - `SFX_Promote_Chime`: Ascending dual-tone arpeggio when object reaches Survivor/Tenured.
  - `SFX_STW_Pause`: Deep low-pass resonant rumble during Stop-The-World pause.
  - `SFX_OOM_Alarm`: Klaxon emergency warning when Old Gen reaches >85%.

---

## 6. Game State Sequence & HUD Layout
- **Scene Flow**: `Boot` (JVM Startup) $\rightarrow$ `Title Screen` $\rightarrow$ `Gameplay` (Waves 1-10) $\rightarrow$ `Victory` (Benchmark Passed) / `Game Over` (OOM Crash).
- **HUD & UI Overlay (960x540)**:
  - **Top Bar**: Heap Telemetry (Eden %, Survivor %, Old Gen % gauge), CPU Cycles balance, Current Wave (e.g. Wave 3/10), Allocation Rate (MB/s).
  - **Bottom Dock**: Tower Selection Bar (Serial $50, Parallel $120, CMS $200, G1 $350, ZGC $500), STW Panic Button (with cooldown).
  - **Center-Right**: Selected Tower Upgrade/Sell Inspector panel.

---

## 7. Technical Scope & Architecture Notes
- **Ebitengine Subsystem Plan (`internal/`)**:
  - `internal/game`: Main `ebiten.Game` implementation, FSM scene manager.
  - `internal/heap`: Grid model, memory lanes (Eden, S0, S1, Old Gen), path waypoints.
  - `internal/entity`: Objects (enemies), Towers (collectors), Projectiles/Beams, Particle pool.
  - `internal/audio`: Pure-code DSP synthesizer, SFX triggers.
  - `internal/ui`: HUD rendering, 9-slice boxes, buttons, telemetry bars.
- **Fail-Fast Compliance**:
  - Strictly zero fallbacks, rich diagnostic error returns, clean modular Go packages.
