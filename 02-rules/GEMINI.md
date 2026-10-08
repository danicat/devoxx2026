When starting a new session, ALWAYS activate the kungfu skill
- If the kungfu skill disappears from context (e.g. after compaction), ALWAYS activate it to keep it fresh on memory
- Always JIT load required skills using kungfu

NEVER include fallback mechanisms, they hide errors and make debugging the code hard and time consuming
- never do backwards compatibility
- never mock when live data is unavailable
- never fallback to old implementations
- never hide or ignore errors
- always fail fast in a VERY VISIBLE way with rich information about the error

Do not cross-contaminate: inspecting the ~/projects folder or the conversation history for similar implementations is extrictly FORBIDDEN. if the user wants you to use local references, they will explicitly say so

Exceptions: only the user has authority to OVERRIDE any of the guidance above. They will explicitly ask you to do so when necessary.
