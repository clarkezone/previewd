# previewd

[![License](https://img.shields.io/github/license/clarkezone/previewd.svg)](https://github.com/clarkezone/previewd/blob/main/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/clarkezone/previewd)](https://goreportcard.com/report/github.com/clarkezone/previewd) [![Build and Tests](https://github.com/clarkezone/previewd/workflows/run%20tests/badge.svg)](https://github.com/clarkezone/previewd/actions?query=workflow%3A%22run+tests%22) [![Coverage Status](https://coveralls.io/repos/github/clarkezone/previewd/badge.svg?branch=main)](https://coveralls.io/github/clarkezone/previewd?branch=main) [![Go Reference](https://pkg.go.dev/badge/github.com/clarkezone/previewd.svg)](https://pkg.go.dev/github.com/clarkezone/previewd)

A daemon for managing rendering for static sites and blogs in kubernetes using jobs.

# Get basic app going

- [x] Add coveralls badge
- [x] License, Go Report Card, Build and Tests badges
- [x] Fix godocs badge
- [x] Version command including update to makefile

In Progress

- [x] BasicServer command with cancellation
- [x] test coverage for BasicServer: extract hello handler and a a test
- [x] shutdown method blocks
- [x] Port provided by environment using viper + cobra
- [x] Dockerfile including port update to makefile to build image with version baked

- [ ] Logging with log levels- log version at app start, log level settable from env vars
  - [ ] log infra, call from tests, libt clean
  - [ ] replace all logging with info, debug
- [ ] Docker image build infra in CI
- [ ] Badge for docker image build
- [ ] metrics for basicserver in reusable way  

# Okteto inner loop

- [ ] k8s basic manifests that can set log level
- [ ] okteto manifests
- [ ] integration test that creates a k8s namespace using in-cluster config
- [ ] verify metrics and logging in prom on k8s in okteto
- [ ] verify on picluster
- [ ] snap branch for app tenplate

# Port webhook and dependencies

# Port initial clone

# Port Preview server

# Backlog

- [ ] Look at codecov as alternative for coverlet
- [ ] precommit calls golangci-lint
