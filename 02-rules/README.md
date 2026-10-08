# 02 - Rules: Keeping Guardrails Lean

Walkthrough of a minimal, production-grade rules file: [GEMINI.md](./GEMINI.md).

## Walkthrough Steps

1. **Open [GEMINI.md](./GEMINI.md)**:
   - Notice it is **under 25 lines** instead of a 500-line monolithic rulebook.
2. **Show the 3 Non-Negotiable Rules**:
   - **Rule 1 (Dynamic Skills)**: Always activate and keep the `kungfu` skill alive.
   - **Rule 2 (Fail Fast Visibly)**: Never use fallbacks, never mock live data, never hide errors.
   - **Rule 3 (Boundary Guards)**: Never inspect or touch other projects in `~/projects`.
3. **Key Point**:
   - Linters, formatters, and type-checks are enforced via external scripts and CI, not begged for in prose prompts.
