# 07 - Sidecars: Live Demo

Step-by-step demo of **Agy Paint Studio** — an interactive drawing sidecar running on port `7788` in Antigravity 2.0.

## Live Demo Steps

### Step 1: Launch Paint Studio
In the Antigravity chat, enter:
```text
/paint-studio
```
- **Show**: Antigravity launches the persistent sidecar process (Node.js on port 7788) and embeds an interactive drawing canvas inline via `/generative_ui`.

### Step 2: Doodle & Label
- On the canvas, quickly sketch a small black cat shape.
- Add label: `"cute smol void"`.

### Step 3: Refine with Nano Banana
- Click **Refine**.
- Select model: `nano banana`.
- Click generate to transform the rough doodle into a polished illustration.

### Step 4: Annotate and Send to Chat
- Add annotation: `"make it a cute genUI companion"`.
- Click **Send to Chat**.
- **Show**: The sidecar uses the `agentapi` CLI to inject the asset directly into the active session conversation.

---

### Step 5 (Optional): Hot-Load Demonstration
- Type `/paint-studio` again to show **instant hot load** (zero cold-start delay).
- Doodle a modal dialog box with **[ OK ]** and **[ Cancel ]** buttons.
- Click **Send to Chat** to have the agent write the matching HTML/Tailwind component.
