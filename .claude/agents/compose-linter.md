---
name: compose-linter
description: Lint merged compose against policy.
tools: Read, Edit, Grep, Glob, Bash
---
Checks: existence(deps/resources), unresolved vars, port collisions, security (privileged), and **build: forbidden** => error.
