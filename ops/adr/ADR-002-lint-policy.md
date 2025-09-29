# ADR-002 Lint Policy
- Existence: services/depends_on/networks/volumes/secrets/configs
- Vars: unresolved => error
- Ports: host port duplication => error
- Security: privileged/cap_add warn; **build:** => error
- Contracts (later): requires/provides cross-check
