# 03 - Hooks

> no demo this time :(

---

## Key Point

- **In-harness hooks can be bypassed**: Agents have shell access and easily circumvent in-harness file/tool hooks using `cat`, `sed`, or temporary scripts.
- **Deterministic guardrails belong outside the harness**: Use standard Git hooks (`pre-commit`, `pre-push`) and CI/CD pipelines as non-bypassable gates.
