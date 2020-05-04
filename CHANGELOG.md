# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.1] - 2020-05-04
### Changed
- DNS and hardcoded seed peers are now whitelisted
- The `ddrpcli add-peer` command now accepts a `--verify` argument to force verifiction of the remote peer ID

## [0.1.1] - 2020-5-04
### Added
- Added `ddrpcli version` and `ddrpd version` CLI commands

### Fixed
- Fixed build scripts to properly insert Git tags and commits at build-time
- Fixed user-agent string