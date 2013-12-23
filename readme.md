# Docket

Docket provides transports and straightforward configuration files for [Docker](https://www.docker.io/).<br/>
Strongly version your containers, then distribute them offline or over SSH & HTTP with git!

Docker's image storage is treated like a cache, while Docket manages your images using git as a persistent storage backend.
Your containers now have effortless history, strong hashes to verify integrity, git commit messages, and secure transport.

Further, ditch those long config flags and express them in a file instead.
Docket looks for a `docket.toml` file in the current directory and sets up binds, mounts, etc.
Add that file to your project's version control and get your entire team working on the same system.

## Quickstart

Grab the latest [release](https://github.com/polydawn/docket/releases) and throw it on your path. Alternately, [build Docket from source](#building-from-source).

```bash
# Clone down some example config files
git clone https://github.com/polydawn/boxen.git && cd boxen/ubuntu

# Download ubuntu from public index, save into git
docket build -s index -d graph --noop

# Upgrade apt-get packages
docket build

# Load repeatable ubuntu from git and start an interactive shell
docket run bash
```

You just built a git repo tracking ubuntu.

A new (bare) repository called `graph` has appeared in the `boxen` folder.<br/>
Push that anywhere and have your team clone it down!

## How do I use it?

Docket provides two commands to help you maintain & use images: `build` and `run`.

Using `build` will take an image from somewhere, execute a build step, and save the result.<br/>
Using `run` just runs an (already-built) image.

### First steps

To use Docket, you need a config file in the current directory called `docket.toml`.<br/>
This file tracks all your image names and settings.

Our [Boxen](https://github.com/polydawn/boxen) repository has several examples, which we'll use for this tutorial.<br/>
Clone it down if you haven't already:

```bash
# Clone down some prepared examples
git clone https://github.com/polydawn/boxen.git && cd boxen
```

In the boxen folder, there's the first [config file](https://github.com/polydawn/boxen/blob/master/docket.toml).
You'll notice there's only one section - `settings`.
Here we set up a bunch of settings we want for pretty much every image: DNS servers, folder mounts, etc.

Because Docket is smart, these settings apply to every image configured in Boxen.
Docket scans up parent folders, looking for `docket.toml` files, and stops when it can't find one.
Today, we'll be using ubuntu:

```bash
# Pop into the ubuntu folder
cd ubuntu
```

You'll notice this next [config file](https://github.com/polydawn/boxen/blob/master/ubuntu/docket.toml) is different - it has *image* and *target* sections, with copious comments.<br/>
We'll explain each in turn:

### Image names

The image section has three entries: *name*, *upstream*, and *index*.

<table>
	<tr>
		<th>Entry</th>
		<th>Purpose</th>
	</tr><tr>
	<tr>
		<td>Name</td>
		<td>
			<p>The name of the image, and the name of the branch that ends up in git.</p>
			<p>Example: <code>polydawn.net/ubuntu/12.04</code></p>
		</td>
	</tr>
	<tr><tr>
		<td>Upstream</td>
		<td>
			<p>This image's parent - where did it get built from? This is used by the build command.</p>

			<p>Example: <code>index.docker.io/ubuntu/12.04</code>
		</td>
	</tr><tr>
	<tr>
		<td>Index</td>
		<td>
			<p><a href="https://index.docker.io">Docker index</a> names are not compatible with reasonable branch names.</p>

			<p>A prime example is "ubuntu", which would be confusing without specifying the difference between an unmodified ubuntu and one you've saved in the graph.</p>

			<p>For this reason, we store the upstream image's index alias separately, and <i>only</i> use that when pulling from the index.</p>

			<p>Example: <code>ubuntu:12.04</code></p>
		</td>
	</tr>
</table>

### Targets

Targets tell Docket what to do.
They can be called anything, but two are special: `build` and `run`, which are the defaults used when you tell Docket to... build & run!

Unlike the settings section, putting settings in a *target* only applies to that target.
It does not affect other folders.

You can put any setting in a target, but the most common usage is to set a different **command**. You'll notice that in the current folder, trying `docket run` will just echo out an example message, while `docket run bash` will launch a bash shell.

Of course, neither will work right now - Docket can't find your image!
We need to get ourselves an image.

### Bootstrapping from the index

Most of the time, you'll want to fork & version an image from the public docker index.
We support plain tarballs as well (more on that later), but the index can be convenient.
This is what the ubuntu [conf file](https://github.com/polydawn/boxen/blob/master/ubuntu/docket.toml) is ready to do.

Try the following:

```bash
# Download ubuntu from public index, save into git
docket build -s index -d graph --noop
```

This command accomplished a few things:

* Docket chose the public index as the *source*, and looked there for an image called `ubuntu:12.04`
* The `--noop` flag means we are just moving images around, so there was no build step. More on that later.
* Once downloaded, Docket saved that image to the graph *destination*.
 * Odds are you didn't have a graph repository, so Docket created one for you.

You now have a (bare) git repository called `graph` in the `boxen` folder!<br/>
If you check out the log, you'll have a single commit with the image's branch name:

```
$ ( cd ../graph ; git log --graph --decorate )

* commit 7105d5622bf8118af1c13001f2b36d51a93f020e (index.docker.io/ubuntu/12.04)
  Author: Your Name <you@example.com>

      index.docker.io/ubuntu/12.04 imported from an external source
```

You can push this repository anywhere & share it with the world.
Whoever receives it can validate the hash and have a guarantee it's the same image.

### Sources & destinations:

Our example used the index as a source and a local graph as the destination.
Docket supports a few others:

<table>
<tr>
	<tr>
		<th>Type</th>
		<th>Purpose</th>
	</tr><tr>
	<tr>
		<td>Graph</td>
		<td>A git repository used to version images. <i>(default)</i></td>
	</tr><tr>
	<tr>
		<td>File</td>
		<td>A tarball created from docker export.</td>
	</tr><tr>
	<tr>
		<td>Docker</td>
		<td>The local docker daemon's cache.</td>
	</tr><tr>
	<tr>
		<td>Index</td>
		<td>The <a href="https://index.docker.io">public index</a>.</td>
	</tr>
</table>

This makes it easy to load & save images in a variety of ways.
Use the appropriate strategy for your situation.<br/>
The default is `graph`; we added `-d graph` to the bootstrapping command to avoid confusion.

### Building an image

We're now ready to fork the image we downloaded and walk our own (strongly-versioned) path.

```bash
# Upgrade apt-get packages & save the new ubuntu image
docket build
```

This will plug away for awhile (you're updating all the ubuntu packages!) and accomplish a few things:

* Docket imported the image from the graph.
 * The Docker daemon now knows about `index.docker.io/ubuntu`, not just `ubuntu`.
* Since there was no `--noop` flag this time, we ran the build step.
 * Check out the [build.sh] file in the current folder: it runs a couple scripts around apt-get.
* After building, Docket saved our new image to the graph.
 * We now have a new branch name, starting with `example.com/ubuntu`.

Your git log has a new commit listed:

```
$ ( cd ../graph ; git log --graph --decorate )

* commit 2a9c8a28220717790de7336d07f86e9857074509 (HEAD, example.com/ubuntu/12.04)
| Author: Your Name <you@example.com>
|
|     example.com/ubuntu/12.04 updated from index.docker.io/ubuntu/12.04
|
* commit 7105d5622bf8118af1c13001f2b36d51a93f020e (index.docker.io/ubuntu/12.04)
  Author: Your Name <you@example.com>

      index.docker.io/ubuntu/12.04 imported from an external source
```

Notice how you now have two branches, named after their respective images.
This git repository will track which image was built from where, using merges - an audit log, built into the log graph.

Now you can play around with docket images. Launch a bash shell and experiment!

```bash
# Load repeatable ubuntu from git and start an interactive shell
docket run bash
```

### What's next?

From here, we strongly recommend playing around more with the example [Boxen](https://github.com/polydawn/boxen) folders.
There's several images pre-configured there, for example a zero-config nginx server.
Additions to that repository are welcome!

When you're ready to use Docket with your own team, simply write your own `docket.toml` file and place it in a new folder.
Build yourself an image (perhaps copying one of our `build.sh` scripts?) and share your machine with the world!


## Building from source

To build Docket, you will need Go 1.2 or newer.
Following the [golang instructions](http://golang.org/doc/install#bsd_linux) for 64-bit linux:

```bash
curl https://go.googlecode.com/files/go1.2.linux-amd64.tar.gz -o golang.tar.gz
sudo tar -C /usr/local -xzf golang.tar.gz
export PATH=$PATH:/usr/local/go/bin # Add this to /etc/profile or similar
```

Clone down Docket & throw it on your path:
```bash
git clone https://github.com/polydawn/docket && cd docket
git submodule update --init

./goad build
sudo cp docket/docket /usr/bin/docket
```

Now you're ready to rock & roll.
Lots of examples are available over at [Boxen](https://github.com/polydawn/boxen)!


## Installing Docker

Docket uses [Docker](https://www.docker.io/), an excellent container helper based on LXC.
This gives Docket all that containerization mojo. We're using Docker 0.7.2 right now.

Docker offers a variety of [installation instructions](http://docs.docker.io/en/latest/installation/) on their site.
