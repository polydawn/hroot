# Docket

Docket makes [Docker](https://www.docker.io/) easy.

Use straightforward configuration files & git to construct, version, and distribute docker images.

## Quickstart

Coming soon!

## How do I use it?

Docket configuration is recursive; each folder is a refinement of its parent.
This allows you to express a complex, structured ecosystem in a natural way.
[Boxen](https://github.com/polydawn/boxen) shows this off a bit by arranging example Docket files for various popular services.

Configuration is split into targets, so changing from debug to production is a breeze. Check out an [example file](https://github.com/polydawn/boxen/blob/master/docker.toml).

Commands:
* **run** *target* launches a docker image.
* **export** *target* will run a target.
* **unpack** *image* will load an image from the graph repository into Docker.
* **publish** *target* will update and image from the graph, then publish the changes.

## Building from source

To build Docket, you will need Go 1.1 or newer.
Following the [golang instructions](http://golang.org/doc/install#bsd_linux) for 64-bit linux:

```bash
curl https://go.googlecode.com/files/go1.1.2.linux-amd64.tar.gz -o golang.tar.gz
sudo tar -C /usr/local -xzf golang.tar.gz
export PATH=$PATH:/usr/local/go/bin # Add this to /etc/profile or similar
```

Clone down Docket & throw it on your path:
```bash
git clone https://github.com/polydawn/docket --recursive
docket/build.sh
sudo cp docket/docket /usr/bin/docket
```

Now you're ready to rock & roll.
Lots of examples are available over at [Boxen](https://github.com/polydawn/boxen)!

## Installing Docker

Docket uses [Docker](https://www.docker.io/), an excellent container helper based on LXC.
This gives Docket all that containerization mojo.
On Ubuntu 13.04, using the latest packaged installation (0.6.x) works fine. From their [instructions](http://docs.docker.io/en/latest/installation/ubuntulinux/):

```bash
sudo apt-get update
sudo apt-get install linux-image-extra-`uname -r`
sudo sh -c "wget -qO- https://get.docker.io/gpg | apt-key add -"
sudo sh -c "echo deb http://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list"
sudo apt-get update
sudo apt-get install lxc-docker
```
