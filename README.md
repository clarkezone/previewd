# previewd

[![License](https://img.shields.io/github/license/clarkezone/previewd.svg)](https://github.com/clarkezone/previewd/blob/main/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/clarkezone/previewd)](https://goreportcard.com/report/github.com/clarkezone/previewd) [![Build and Tests](https://github.com/clarkezone/previewd/workflows/run%20tests/badge.svg)](https://github.com/clarkezone/previewd/actions?query=workflow%3A%22run+tests%22) [![Coverage Status](https://coveralls.io/repos/github/clarkezone/previewd/badge.svg?branch=main)](https://coveralls.io/github/clarkezone/previewd?branch=main) [![Go Reference](https://pkg.go.dev/badge/github.com/clarkezone/previewd.svg)](https://pkg.go.dev/github.com/clarkezone/previewd)

A daemon for managing rendering for static sites and blogs in kubernetes using jobs.

# ~~Get basic skeleton app going~~

- [x] Badges, CI/CD, Test infra, Code Coverage, license, Linting, precommit, Dockfile, basic cli app with test server with basic logging and metrics

# Okteto inner loop

- [x] k8s basic manifests that can set log level, verify on picluster, okteto manifests

# Port webhook and dependencies

- [ ] port all packages
  - [x] JobManager
  - [x] Kubelayer
  - [x] local resource manager
  - [x] port lrm to new logging
  - [x] port kubelayer to new logging
  - [x] port webhook listener
  - [ ] add logging and metrics to listener using middleware
    - [x] workout how to bucket counters by successtype
    - [x] Fix endpoint in duration
    - [ ] parameterize metrics names in middleware
    - [ ] Fix /metrics in testserver to hook in basicserver
    - [ ] Add logging and metrics to webhook
    - [ ] ensure test coverage for middleware
- [ ] integration test that calls webhook job creation code that uses out of cluster mode based on main function
- [ ] update integration test to work in okteto with auto-detect
- [ ] make command to call integration test
- [ ] github action to configure okteto connection and call integration test
- [ ] can code coverage reflect integration test
- [ ] verify metrics and logging in prom on k8s in okteto

# Port initial clone

- [ ] port main function into cobra commands

# Port Preview server

# Backlog

- [ ] Badge for docker image build
- [ ] Look at codecov as alternative for coverlet
- [ ] precommit calls golangci-lint
