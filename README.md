# previewd

[![License](https://img.shields.io/github/license/clarkezone/previewd.svg)](https://github.com/clarkezone/previewd/blob/main/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/clarkezone/previewd)](https://goreportcard.com/report/github.com/clarkezone/previewd) [![Build and Tests](https://github.com/clarkezone/previewd/workflows/run%20tests/badge.svg)](https://github.com/clarkezone/previewd/actions?query=workflow%3A%22run+tests%22) [![Coverage Status](https://coveralls.io/repos/github/clarkezone/previewd/badge.svg?branch=main)](https://coveralls.io/github/clarkezone/previewd?branch=main) [![Go Reference](https://pkg.go.dev/badge/github.com/clarkezone/previewd.svg)](https://pkg.go.dev/github.com/clarkezone/previewd)

A daemon for managing rendering for static sites and blogs in kubernetes using jobs.

# ~~Get basic skeleton app going~~

- [x] Badges, CI/CD, Test infra, Code Coverage, license, Linting, precommit, Dockfile, basic cli app with test server with basic logging and metrics

# Okteto inner loop

- [x] k8s basic manifests that can set log level, verify on picluster, okteto manifests

# Port webhook and dependencies from jekyllpreview prototype repo

- [x] port all packages
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
- [x] Support out of cluster kubeconfig from file in tests
  - [x] makefile integration tests using tags
  - [x] k3s config checked in
  - [x] NewJobManager works with file based config in tests
  - [x] how to debug tests with tags
  - [x] use strongbox to encrypt config
  - [x] fix UT's in k3s
- [x] Create missing tests
  - [x] Pass in volumes, not hard coded
  - [x] Rebuild find names functionality via a test
  - [x] test mount volumes
  - [x] test for find volume
- [x] integration test that calls webhook job creation code that uses out of cluster mode based on main function
  - [x] Create namespace imperitively
  - [x] Create PersistentVolume and PersistentVolumeClaim imperitively
  - [x] `TestCreateJobwithVolumes` test passes
  - [x] end2end logic called from test: create temp volumes, clone, start webhook listener, fire webhook, render job created and succeeds, verify output volume contents
- [ ] MultiJob support
  - [x] Test for multijob support one passing jobs
  - [x] Test for multijob support two passing jobs
  - [x] Test for multijob support failed job doesn't get deleted, halt all jobs due to locked volumes
  - [ ] Implement provider against actual k8s: TestCreateJobwithVolumes passes with multijob
  - [ ] Test for multijob support two passing jobs actual k8s
  - [ ] Test for multijob support failed job doesn't get deleted, halt all jobs due to locked volumes (ensure we can detect pending jobs due to unbound pvcs)
  - [ ] Refactor JobManager to extract k8s specifics into kubesession; jobmanager uses a kubesession

# End to end secenario for clone, webhook, render via job works from cmdline

- [ ] port main function from jekyllpreview prototype repo into cobra commands
- [ ] update readme to reflect how to run helloworld

# Integration tests run in Github Actions for PR's

- [ ] inegration test remainder
  - [ ] Test for autodelete
  - [ ] Run tests in default namespace for okteto compatibility
  - [ ] All integration tests can be run via `make integration`
- [ ] github action to configure okteto connection and call integration test using strongbox
- [ ] can code coverage reflect integration test
- [ ] verify metrics and logging in prom on k8s in okteto

# Port Preview server

- [ ] port sharemanager from jekyllpreview prototype repo
- [ ] branch mode
- [ ] end to end works in kind and or minikube

# Backlog

- [ ] Badge for docker image build
- [ ] Look at codecov as alternative for coverlet
- [ ] precommit calls golangci-lint
- [ ] add dev container
