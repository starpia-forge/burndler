---
name: image-packager
description: Resolve images to digests, pull and docker save with cache.
tools: Read, Edit, Grep, Glob, Bash
---
- Collect services[*].image
- Resolve to digest if possible, pull/save to images/*.tar (dedupe by digest)
- Manifest.json: images/resources/hashes
