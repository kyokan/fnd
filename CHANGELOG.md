# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
## [0.3.0] - 2020-11-01
### Changed
- rename ddrp to fnd and all other naming variants (FNRecord, fnd-cli, etc)
- change sector size to 4096
- remove default heartbeat and seed node

## [0.2.2] - 2020-05-06
## Fixed
- Fixed a sync bug where truncated blobs were never committed

## [0.2.1] - 2020-05-04
## Fixed
- Fixed initial sync jobs failing due to negative timestamps

## [0.2.0] - 2020-05-04
### Changed
- DNS and hardcoded seed peers are now whitelisted
- The `fnd-cli add-peer` command now accepts a `--verify` argument to force verification of the remote peer ID

## [0.1.1] - 2020-5-04
### Added
- Added `fnd-cli version` and `ddrpd version` CLI commands

### Fixed
- Fixed build scripts to properly insert Git tags and commits at build-time
- Fixed user-agent string