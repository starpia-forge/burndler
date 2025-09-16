/**
 * Semantic Release Configuration for Burndler
 *
 * This configuration enables automatic versioning and release management
 * based on conventional commit messages.
 *
 * Commit message format:
 * - feat: new feature (minor version bump)
 * - fix: bug fix (patch version bump)
 * - BREAKING CHANGE: major version bump
 * - chore, docs, style, refactor, test: no version bump
 */

module.exports = {
  branches: [
    'main',
    {
      name: 'develop',
      prerelease: 'beta'
    }
  ],
  plugins: [
    [
      '@semantic-release/commit-analyzer',
      {
        preset: 'conventionalcommits',
        releaseRules: [
          { type: 'feat', release: 'minor' },
          { type: 'fix', release: 'patch' },
          { type: 'perf', release: 'patch' },
          { type: 'revert', release: 'patch' },
          { type: 'docs', release: false },
          { type: 'style', release: false },
          { type: 'chore', release: false },
          { type: 'refactor', release: false },
          { type: 'test', release: false },
          { scope: 'no-release', release: false }
        ],
        parserOpts: {
          noteKeywords: ['BREAKING CHANGE', 'BREAKING CHANGES']
        }
      }
    ],
    [
      '@semantic-release/release-notes-generator',
      {
        preset: 'conventionalcommits',
        presetConfig: {
          types: [
            { type: 'feat', section: '🚀 Features' },
            { type: 'fix', section: '🐛 Bug Fixes' },
            { type: 'perf', section: '⚡ Performance Improvements' },
            { type: 'revert', section: '⏪ Reverts' },
            { type: 'docs', section: '📚 Documentation', hidden: false },
            { type: 'style', section: '💎 Styles', hidden: true },
            { type: 'chore', section: '🧹 Chores', hidden: true },
            { type: 'refactor', section: '♻️ Code Refactoring', hidden: false },
            { type: 'test', section: '✅ Tests', hidden: true },
            { type: 'build', section: '🏗️ Build System', hidden: false },
            { type: 'ci', section: '👷 CI/CD', hidden: false }
          ]
        }
      }
    ],
    [
      '@semantic-release/changelog',
      {
        changelogFile: 'CHANGELOG.md',
        changelogTitle: '# Changelog\n\nAll notable changes to this project will be documented in this file.'
      }
    ],
    [
      '@semantic-release/npm',
      {
        npmPublish: false // We don't publish NPM packages
      }
    ],
    [
      '@semantic-release/exec',
      {
        prepareCmd: 'echo "${nextRelease.version}" > VERSION && make build-docker VERSION=${nextRelease.version}',
        publishCmd: 'echo "Release ${nextRelease.version} completed"'
      }
    ],
    [
      '@semantic-release/github',
      {
        assets: [
          {
            path: 'dist/burndler',
            name: 'burndler-${nextRelease.gitTag}-linux-amd64',
            label: 'Burndler Binary (Linux AMD64)'
          },
          {
            path: 'dist/burndler-merge',
            name: 'burndler-merge-${nextRelease.gitTag}-linux-amd64',
            label: 'Burndler Merge Tool (Linux AMD64)'
          },
          {
            path: 'dist/burndler-lint',
            name: 'burndler-lint-${nextRelease.gitTag}-linux-amd64',
            label: 'Burndler Lint Tool (Linux AMD64)'
          },
          {
            path: 'dist/burndler-package',
            name: 'burndler-package-${nextRelease.gitTag}-linux-amd64',
            label: 'Burndler Package Tool (Linux AMD64)'
          }
        ]
      }
    ],
    [
      '@semantic-release/git',
      {
        assets: ['CHANGELOG.md', 'VERSION', 'frontend/package.json'],
        message: 'chore(release): ${nextRelease.version} [skip ci]\n\n${nextRelease.notes}'
      }
    ]
  ]
};