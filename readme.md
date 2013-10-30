# Docket

Docket makes [Docker](https://www.docker.io/) easy.

Use straightforward configuration files & git to construct, version, and distribute docker images.

## Quickstart

Grab the latest [release](https://github.com/polydawn/docket/releases) and throw it on your path.

*This example does not work yet, we're working on it!*
```bash
# Clone down some example config files
git clone https://github.com/polydawn/boxen.git && cd boxen

# Create a graph repo
git init graph && cd graph
git commit -m "Graph" --allow-empty && cd ../ubuntu

# Download ubuntu from public index
docket build index -s index -d graph --noop

# Upgrade apt-get packages
docket build index

# Load repeatable ubuntu from git and start an interactive shell
docket run bash
```

Now you've got a git repository tracking ubuntu!

## How do I use it?

Docket configuration is recursive; each folder is a refinement of its parent.
This allows you to express a complex, structured ecosystem in a natural way.
[Boxen](https://github.com/polydawn/boxen) shows this off a bit by arranging example Docket files for various popular services.

Configuration is split into targets, so changing from debug to production is a breeze. Check out an [example file](https://github.com/polydawn/boxen/blob/master/docker.toml).

### Commands:
<table>
	<tr>
		<td>run</td>
		<td>Launch a docker image.</td>
	</tr><tr>
		<td>build</td>
		<td>Do the same thing, then save the results somewhere.</td>
	</tr>
</table>

### Sources & destinations:
<table>
	<tr>
		<td>graph</td>
		<td>A git repository used to version images <i>(default)</i></td>
	</tr><tr>
	<tr>
		<td>file</td>
		<td>A tarball created from docker export</td>
	</tr><tr>
	<tr>
		<td>docker</td>
		<td>The docker daemon</td>
	</tr><tr>
	<tr>
		<td>index</td>
		<td>The <a href="https://index.docker.io">public index</a></td>
	</tr>
</table>

This makes it easy to load & save images in a variety of ways.

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
