# previewd

[![License](https://img.shields.io/github/license/clarkezone/previewd.svg)](https://github.com/clarkezone/previewd/blob/main/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/clarkezone/previewd)](https://goreportcard.com/report/github.com/clarkezone/previewd) [![Build and Tests](https://github.com/clarkezone/previewd/workflows/run%20tests/badge.svg)](https://github.com/clarkezone/previewd/actions?query=workflow%3A%22run+tests%22) [![Coverage Status](https://coveralls.io/repos/github/clarkezone/previewd/badge.svg?branch=main)](https://coveralls.io/github/clarkezone/previewd?branch=main) [![Go Reference](https://pkg.go.dev/badge/github.com/clarkezone/previewd.svg)](https://pkg.go.dev/github.com/clarkezone/previewd)

## Description

`previewd` is a daemon that is primarily designed to be deployed into a kubernetes cluster to facilitate previewing and hosting of static websites built using a static site generator such as Jekyll, Hugo or Publish.

```mermaid
graph  LR
  client([browser])-..->ingress[Ingress];
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

## Installation and use

This project is still in development and as such we don't yet have instructions for how to use. That said you can build the code and run the tests. The backlog is maintained in [docs/workbacklog.md](docs/workbacklog.md)

### Dev Setup

1. Install a recent version of golang, we recommend 1.17 or greater.
2. In order to run integration tests you'll need a kubernetes environment. For development we use k3s
3. Update `GetTestConfigPath()` to point to a valid kubeconfig for your test cluster
4. Define environment variables TODO
5. precommit TODO
6. githook TODO

### Build from Source

To compile the main previewd executable:

```bash
make build
```

Which will produce a binary at <projectroot>/bin/previewd

To build the docker container

```bash
docker build .
```

### Run tests

This project comes with both unit and integration tests. Unit tests run standalone with no dependencies other than test environment configuration. Integration tests require a kubernetes cluster.

For unit tests, use:

```bash
make test
```

For integration test, use:

```bash
make integration
```
