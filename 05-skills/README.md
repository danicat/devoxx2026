# 05 - Agent Skills: Live Demo

Step-by-step demo of **KungFu** Just-In-Time (JIT) skill discovery and prompt injection.

Supporting prompt: [PROMPT.md](./PROMPT.md)

---

## Live Demo Steps

### Part 1: Terminal CLI Walkthrough
Run these in the terminal to demonstrate skill discovery and catalog sync:

1. **Find skills by keyword** (hybrid BM25 + TF-IDF search):
   ```bash
   kungfu find game
   ```
2. **List available & cached skills**:
   ```bash
   kungfu list
   ```
3. **Synchronize catalogs & check updates**:
   ```bash
   kungfu catalog sync
   kungfu update
   ```

4. **JIT loading**:
   ```bash
   kungfu load kungfu
   ```

### Part 2: In-Agent Prompting (JIT Injection)
Switch to the Antigravity session and paste the demo prompt from [PROMPT.md](./PROMPT.md):