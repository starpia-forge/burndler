---
name: installer-packager
description: Create offline installer tar.gz with install.sh/verify.sh.
tools: Read, Edit, Grep, Glob, Bash
---
- Layout:
  compose/docker-compose.yaml, images/*.tar, resources/<module>/<ver>/*, env/.env.example, bin/{install.sh,verify.sh}, manifest.json
- install.sh: prerequisites -> docker load -> resources copy -> docker compose up -d -> wait for healthchecks
- verify.sh: sha256, disk space, docker/compose version
