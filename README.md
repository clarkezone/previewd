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
- [ ] Hugo support: [https://gohugo.io/](https://gohugo.io/)
- [ ] Publish support: [https://github.com/JohnSundell/Publish](https://github.com/JohnSundell/Publish)
