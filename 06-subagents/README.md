# 06 - Subagents: Live Demo

Step-by-step demo of dynamic swarm orchestration and context isolation ("coordinator as tech lead").

## Pedantic Open Source Review Swarm

Fans out 10 reviewer agents to audit a codebase and initiates an interactive `/grill-me` alignment.

1. Run [PROMPT.md](./PROMPT.md) in agy CLI with @PROMPT.md

2. **What to Show**:
   - `::uno-reverse`: Adversarial reviewer persona.
   - `::parallel(10)`: 10 worker subagents audit separate files simultaneously.
   - `/grill-me`: Coordinator gathers findings and grills the developer on trade-offs.
