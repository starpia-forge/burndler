# Changelog

All notable changes to this project will be documented in this file.

## 1.0.0 (2025-09-18)

### 🚀 Features

* add `build-docker` target to Makefile for Docker image generation ([80f00c7](https://github.com/starpia-forge/burndler/commit/80f00c750d130645a1720739f8df0e11c9a255a7))
* add Admin role support with expanded functionality and tests ([0470880](https://github.com/starpia-forge/burndler/commit/0470880419c77d77d6f39c4c042e4b034f895b86))
* add password authentication to User model ([6480d34](https://github.com/starpia-forge/burndler/commit/6480d348d36daa508ddac73a577218464d8d0775))
* add production-ready Docker setup with Compose support ([def23da](https://github.com/starpia-forge/burndler/commit/def23daa3e65994f2a60ec5377106d6316cf2adf))
* add setup management with endpoints, models, and middleware ([e76268d](https://github.com/starpia-forge/burndler/commit/e76268db23e3cf81fdceb4756099ee0bdd18831e))
* **build:** add version injection to binaries and Docker builds ([425146c](https://github.com/starpia-forge/burndler/commit/425146c006f8b711098e516ab56e08782b3656fd))
* **ci:** add GitHub Actions CI/CD pipelines for automated testing and releases ([c2d0d26](https://github.com/starpia-forge/burndler/commit/c2d0d269a5436a34281a4decff97e83de75a3d29))
* **deps:** add Dependabot configuration for automated dependency updates ([3dc6b5e](https://github.com/starpia-forge/burndler/commit/3dc6b5e3977ea5479be55ff1e9e03e1a79ac39ab))
* embed frontend assets into backend with SPA routing support ([922b62e](https://github.com/starpia-forge/burndler/commit/922b62e82b8c41052f4f74a0fac4f3b2960bdb08))
* enforce size limits and improve validation for storage backends ([4ff9fee](https://github.com/starpia-forge/burndler/commit/4ff9feeddbf0fc928dc70a7d22f0a9818b106516))
* implement authentication service with JWT ([169aa8f](https://github.com/starpia-forge/burndler/commit/169aa8fcc3bb149f97029566b1b622dfc02382ec))
* implement comprehensive theme system and authentication ([2592190](https://github.com/starpia-forge/burndler/commit/2592190a58ca4eb2f3db41c6ac0483046924aaf4))
* implement login and refresh token handlers ([5db6f0f](https://github.com/starpia-forge/burndler/commit/5db6f0fe5a95181be8e87dad36a1c8323c62d45c))
* integrate authentication routes into server ([33e383e](https://github.com/starpia-forge/burndler/commit/33e383e213cf496e09b32b1715bb094cf42aebd4))
* **make:** add release management targets and version-aware builds ([72e30e0](https://github.com/starpia-forge/burndler/commit/72e30e06ab469ff3a5c7f2fe7d2bc44c3cc01a3f))
* **quality:** add pre-commit hooks and linting configurations ([9a5ab67](https://github.com/starpia-forge/burndler/commit/9a5ab67e01fa7e5edd43043fe30c114a607c7de1))
* **release:** implement semantic versioning with automated changelog ([2a9bfc5](https://github.com/starpia-forge/burndler/commit/2a9bfc59b22eb49357b69737e217ad4fad4f8f26))

### 🐛 Bug Fixes

* add missing newline at EOF in config_test.go ([e40015e](https://github.com/starpia-forge/burndler/commit/e40015e6bd7776c8a6d7634696587ee51683b63a))
* add missing newline at EOF in test files ([f35b5c3](https://github.com/starpia-forge/burndler/commit/f35b5c338a6b32fabeccba5c8c7227c4fff63582))
* add newline at end of files for consistency ([034ac79](https://github.com/starpia-forge/burndler/commit/034ac7907153de2ea4f08f720f0146bb20f92c1d))
* align struct field declarations for consistent formatting ([6a2c4f9](https://github.com/starpia-forge/burndler/commit/6a2c4f931a43ecd30e5fc20173dd38862fa2332e))
* CI and build systems [#1](https://github.com/starpia-forge/burndler/issues/1) ([b26b7a3](https://github.com/starpia-forge/burndler/commit/b26b7a3be13e06f07f928ed30346ed78c58dc462))
* **ci:** resolve false positive in build directive validation ([00df066](https://github.com/starpia-forge/burndler/commit/00df06662bf15fc9414e6817e166546d66de931a))
* **ci:** update TruffleHog path parameters for PR event compatibility ([d411932](https://github.com/starpia-forge/burndler/commit/d41193218aa82a9fcf9afdad96fc3f2c23a196cf))
* **release:** Fix semantic release workflow failures ([7c64fa2](https://github.com/starpia-forge/burndler/commit/7c64fa213f12456682967f99849d3da60b81d078))
* **release:** Release workflow (Merge pull request [#4](https://github.com/starpia-forge/burndler/issues/4)) ([15fc4e1](https://github.com/starpia-forge/burndler/commit/15fc4e16774599a03be37348804ccedfc539be4e))

### 📚 Documentation

* add CI badges and release management documentation ([ca19499](https://github.com/starpia-forge/burndler/commit/ca19499f91f5b3af95cddd49957dac9ed6c7c21c))

### ♻️ Code Refactoring

* consolidate server lifecycle into app package ([6e43a2c](https://github.com/starpia-forge/burndler/commit/6e43a2c2d725419faa7cc96a64503da593b5bccb))
* extract app initialization and server logic from main package ([49cd36d](https://github.com/starpia-forge/burndler/commit/49cd36de9e3285372519b38f3d911006740a7e18))
* separate server logic from app package ([f29794b](https://github.com/starpia-forge/burndler/commit/f29794b7dd592f5d37e1a02affba1dce68ee7af2))
