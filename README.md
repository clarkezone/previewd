# previewd

[![License](https://img.shields.io/github/license/clarkezone/previewd.svg)](https://github.com/clarkezone/previewd/blob/main/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/clarkezone/previewd)](https://goreportcard.com/report/github.com/clarkezone/previewd) [![Build and Tests](https://github.com/clarkezone/previewd/workflows/run%20tests/badge.svg)](https://github.com/clarkezone/previewd/actions?query=workflow%3A%22run+tests%22) [![Coverage Status](https://coveralls.io/repos/github/clarkezone/previewd/badge.svg?branch=main)](https://coveralls.io/github/clarkezone/previewd?branch=main) [![Go Reference](https://pkg.go.dev/badge/github.com/clarkezone/previewd.svg)](https://pkg.go.dev/github.com/clarkezone/previewd)

Previewd is a daemon that is primarily designed to be deployed into a kubernetes cluster to facilitate previewing and hosting of static websites built using a static site generator such as Jekyll, Hugo or Publish.

```mermaid
graph  LR
  client([client])-..->ingress[Ingress];
  webhook([git webhook])-..->hookingress[webhook receiver<BR>ingress];
  ingress-->service[Service];
  hookingress-->previewdservice
  repo-->webhook
  subgraph git
  repo
  end
  subgraph cluster
  hookingress;
  ingress[Web<BR>frontend<BR>ingress];
  previewdservice-->pod3[previewd<BR>jobmanager pod]
  repo-..pull content..->pod3
  pod3-->renderjob
  renderjob[RenderJob<BR>Jekyll image]
  pod3-->source
  source-->renderjob
  renderjob-->render
  pod1-->render
  pod2-->render
  source[Source<BR>volume]
  render[Render<BR>volume]
  service-->pod1[nginx replica 1];
  service-->pod2[inginx replica 2];

  end
  classDef plain fill:#ddd,stroke:#fff,stroke-width:4px,color:#000;
  classDef k8s fill:#326ce5,stroke:#fff,stroke-width:4px,color:#fff;
  classDef cluster fill:#fff,stroke:#bbb,stroke-width:2px,color:#326ce5;
  classDef volume fill:#fff,stroke:#bbb,stroke-width:2px,color:#326ce5;
  class ingress,hookingress,service,pod1,pod2,pod3,previewdservice, k8s;
  class client plain;
  class cluster cluster;
  class source volume
  class render volume
```

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
- [x] MultiJob support
  - [x] Test for multijob support one passing jobs
  - [x] Test for multijob support two passing jobs
  - [x] Test for multijob support failed job doesn't get deleted, halt all jobs due to locked volumes
  - [x] Implement provider against actual k8s: TestCreateJobwithVolumes passes with multijob
  - [x] Refactor JobManager to extract k8s specifics into kubesession; jobmanager uses a kubesession
  - [x] Test for multijob support two passing jobs actual k8s
- [ ] End to end test for webhook
  - [x] Hook up clone only cmdline option (testprep)
  - [x] Hook up initial render (no clone), webhook arg (e2etest)
  - [x] verify exe
  - [x] Prepare environment with PV populated from repo by creating job using previewd container from CI
  - [ ] Make E2E test work
  - [ ] e2e test works with okteto / minikube / kind in default namespace

# End to end secenario for clone, webhook, render via job works from cmdline

- [x] port main function from jekyllpreview prototype repo into cobra commands
- [ ] test for webhook using curl
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
- [ ] Add support for multiple target repos with multiple webhooks / render destinations
- [ ] Hugo support: [https://gohugo.io/](https://gohugo.io/)
- [ ] Publish support: [https://github.com/JohnSundell/Publish](https://github.com/JohnSundell/Publish)
- [ ] add dev container
- [ ] Test for multijob support failed job (eg due to can't bind PV) doesn't get deleted, halt all jobs due to locked volumes (ensure we can detect pending jobs due to unbound pvcs)
