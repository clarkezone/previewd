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
  - [x] add logging and metrics to listener using middleware
    - [x] workout how to bucket counters by successtype
    - [x] Fix endpoint in duration
    - [x] parameterize metrics names in middleware
    - [x] Fix /metrics in testserver to hook in basicserver
    - [x] Add logging and metrics to webhook
    - [x] ensure test coverage for middleware
- [ ] Support out of cluster kubeconfig from file in tests
  - [x] makefile integration tests using tags
  - [x] k3s config checked in
  - [x] NewJobManager works with file based config in tests
  - [ ] how to debug tests with tags
  - [ ] use strongbox to encrypt config
  - [ ] fix UT's in k3s
    - [ ] Run tests in default namespace for okteto compatibility
    - [ ] Pass in volumes, not hard coded
- [ ] update integration test to work in okteto, using default namespace
- [ ] integration test that calls webhook job creation code that uses out of cluster mode based on main function
- [ ] cobra command hooked up for e2e flow
- [ ] github action to configure okteto connection and call integration test using strongbox
- [ ] can code coverage reflect integration test
- [ ] verify metrics and logging in prom on k8s in okteto

# Port initial clone

- [ ] port main function into cobra commands

# Port Preview server

# Backlog

- [ ] Badge for docker image build
- [ ] Look at codecov as alternative for coverlet
- [ ] precommit calls golangci-lint
