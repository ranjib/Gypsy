## Gypsy

Gypsy will be a CI system that levereges [LXC](http://linuxcontainers.org/) for containment and [nomad](https://nomadproject.io/) for scheduling.

### Description

Gypsy can poll github repositories and build projects using YAML based configuration
file. It can also save the build artifacts. Gypsy uses LXC to containerize the
builds and nomad to schedule them. Gypsy is built to be able to run inside a raspberry pi.
But it should be scalable via nomad. In future I plan to add vitess or cockrochdb backend to
gypsy server for HA.

Gypsy is under heavy development. Currently the http endpoint and LXC bits are in place, together
they can be run as an standalone CI system. Gypsy is heavily influenced by GoCD.


### Prequisites

You will need lxc

```
apt-get install lxc lxc-dev
```

You will need to have your [Golang development environment setup](https://golang.org/doc/code.html#Workspaces).

### Usage

- Build
I am assuming you have golang development environment setup, as well as LXC configured (i use ubuntu 14.04)

```
go get github.com/ranjib/gypsy
cd $GOPATH/src/github.com/ranjib/gypsy
make
```

- Run
```sh
gypsy server
```

- Upload a pipeline
```
curl -X PUT --data-binary @/tmp/telegraf.yml http://localhost:5678/pipelines
```
- Build LXC containers from Dockerfile
```
gypsy dockerfile
```
Theres a lot to do. But this is the MVP :-)
### Architecture

Gypsy has two main components, server and client. Gypsy servers provide http end point to interact with gypsy,
manipulate pipelines(projects), and poll repositories. Gypsy clients does the actual building, using linux
containers. Gypsy server's http api is documented [here](https://github.com/ranjib/Gypsy/tree/master/API.md).

### LICENSE

Gypsy - A nomadic CI system

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
