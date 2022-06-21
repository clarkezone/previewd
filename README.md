# previewd

[![License](https://img.shields.io/github/license/clarkezone/previewd.svg)](https://github.com/clarkezone/previewd/blob/main/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/clarkezone/previewd)](https://goreportcard.com/report/github.com/clarkezone/previewd) [![Build and Tests](https://github.com/clarkezone/previewd/workflows/run%20tests/badge.svg)](https://github.com/clarkezone/previewd/actions?query=workflow%3A%22run+tests%22) [![Coverage Status](https://coveralls.io/repos/github/clarkezone/previewd/badge.svg?branch=main)](https://coveralls.io/github/clarkezone/previewd?branch=main) [![Go Reference](https://pkg.go.dev/badge/github.com/clarkezone/previewd.svg)](https://pkg.go.dev/github.com/clarkezone/previewd)

## Description

`previewd` is a daemon that is designed to be deployed into a kubernetes cluster to facilitate previewing and hosting of static websites built using a static site generator such as Jekyll, Hugo or Publish.

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

Previewd is currently at MVP level of maturity; the basic scenario is working as of release 0.4. Previewd can be deployed into a kubernetes cluster, will clone a static website source from github or gitea hosted site, perform an initial render by scheduling a kubernetes job, and will then listen for webhook triggered by a push to the repo. When the webhook is fired, a rebuild job will be scheduled. The resulting output can be hosted using an instance of Nginx. A set of sample manifests are included in the k8s directory of this repo. Instructions below show how these can be applied to play with the basic scenario.

The backlog is maintained in [docs/workbacklog.md](docs/workbacklog.md). The current focus for the project is building out a production ready set of kubernetes manifests and infrastructure to enable selfhosting of a site leveraging previewd on a home cluster including metrics, monitoring, alerting and high availability. Once that step is complete, work will resume to start tackling the feature backlog.

### Clone a static website and render

1. Apply manifests: `kubectl apply -f .`
2. port-forward the ngnix container: `kubectl port-forward -n previewdtest pod/nginxdeployment-7f5454bbdb-gxc5n 8080:8080 --address=0.0.0.0`
3. Point browser at exposed endpoint to view resulting website

### Trigger webhook

1. Port-forward webhook
2. curl a thing

### Using in production environment

## Development Environment for previewd

### Install Tools

1. Install a recent version of golang, we recommend 1.17 or greater. [https://go.dev/doc/install](https://go.dev/doc/install)
2. ensure that your path includes: `export PATH=$PATH:$HOME/go/bin` so that tools installed with `go install` work correctly
3. Install `make` (debian linux: `sudo apt install make`)
4. Install `gcc` (debian linux: `sudo apt install build-essential`)
5. Install [`pre-commit`](https://pre-commit.com/) (debian linux: `sudo apt install precommit`)
6. If you are planning on submitting a PR to this repo, install the git pre-commit hook (`pre-commit install`)
7. Install [`shellcheck`](https://github.com/koalaman/shellcheck) (`sudo apt install shellcheck`e
8. Install tools other golang based linting tools `make install-tools`
9. If you are planning to use VSCode, ensure you have all of the golang tooling installed

### Install a k3s test cluster (optional, required to run tests and scenarios)

Unless you have access to an existing Kubernetes test cluster with a default storage volume provider,

1. Install [`k3s`](https://github.com/k3s-io/k3s): `curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644`
2. Install a block storage provider such as [`Longhorn`](https://longhorn.io/): `kubectl apply -f https://raw.githubusercontent.com/longhorn/longhorn/v1.3.0/deploy/longhorn.yaml`.
3. You will also need to install support for [rwx-workloads](https://longhorn.io/docs/1.2.4/advanced-resources/rwx-workloads/) on longhorn by way of installing nfs mounting tools: `apt install nfs-common`.
4. Finally, ensure there is only one default storageclass (longhorn): `kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'`

### Dev Setup

1. Configure environment variables in shell. There is a .env file in the scripts directory to establish environment variables needed for unit tests. These need to be applied in your shell and in vscode. You can add the following to your .bashrc or manually.

   ```bash
   export $(cat scripts/.previewd_test.env | xargs)
   ```

2. (optional) if you are planning to debug tests in vscode, you'll need to tell VScode about the environment variables and also enable integration tests. In VS code, go to the command palette and search for `Preferences: Open Remote settings (JSON)` and add the following snippet:

   ```json
   {
     "go.buildFlags": ["-tags=unit,integration"],
     "go.buildTags": "-tags=unit,integration",
     "go.testTags": "-tags=unit,integration"
   }
   ```

   Then search for `User Settings` and add the following snippet for environment variables:

   ```json
   "go.testEnvFile": "/home/james/.previewd_test.env",
   ```

3. Edit `internal/testutils.go` and change the value returned by `GetTestConfigPath()` to point to a valid kubeconfig for your test cluster (eg `~/.kube.config` in typical setups or `/etc/rancher/k3s/k3s.yaml` if you followed the instructions above and installed k3s)

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

For integration test, since these depend on having a valid k8s cluster to work properly, make sure you followed step 3 in the dev setup list above. Then use:

```bash
make integration
```

Finally, for end-to-end tests that exercise the initial MVP

```bash
make end2end
```

TODO: run tests in VS (exports)
