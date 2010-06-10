This directory houses the code and resources for the proof of concept  
version of Ampify — Ampify Zero.

IT DOESN'T WORK YET.

**Quickstart**

Assuming you want to call your instance "kickass", run:

    $ ampzero init kickass
    $ ampzero run kickass

And then, assuming you'd chosen the default port settings, point your  
browser at [https://localhost:8040]. Tada!

**Introduction**

Assuming that you've successfully built all the dependencies according  
to the root [README] file, you should now be all set to run Ampify Zero.

Unfortunately, there are a number of different components involved in an  
Ampify Zero setup (Keyspace cluster, Redis servers, Google App Engine  
instances and Ampnodes).

And unless you're used to modern "high-scalability" deployments in other  
contexts, it could all be quite a head fuck. Therefore, a utility script  
is provided to help ease the pain.

**Architecture Overview**

The Ampify Zero network is built up of independent instances (Hosts)  
and a minimal registry at `amphub.org`. The vision is that, eventually,  
every individual would run their own Host.

A Host can be run from a single machine, e.g. on a mobile device or  
on someone's laptop, or it can be deployed across multiple servers —  
even across multiple datacenters.

Hosts are issued with a unique ID number when they first register  
their public key with `amphub`. They can later use their private  
key to authenticate themselves to other Hosts and use their control  
key to make updates at the `amphub` registry regarding their:

* current location (domain name/ip address + port)
* public key
* X.509 certificate for HTTPS frontends

So, for `Host 23` to talk to `Host 51`, it'd first ask `amphub` for  
Host 51's current location, key and certificate. It'd then use that  
information to establish and verify a connection to Host 51.

Users are also similarly issued with a unique ID number when they  
first register their public key with `amphub`. They can later use  
their control key to make updates at `amphub` regarding their:

* current Host's ID number
* public key

So, for `user 42` to send a message to `user 140`, user 42's Host  
would first look up the current Host for user 140, and then the  
current location for the Host before sending the message to it.

It's important to note that in Ampify, unlike existing systems  
like email and many "decentralised web" initiatives, the user's  
ID isn't tied down to any specific Host.

This important distinction means that users have the freedom  
to move Hosts without inconveniences like telling everyone that  
they now have a new ID at a different Host.

This feature will also allow for some of the funkier Ampify 0.x  
functionality, e.g. being able to specify multiple current Hosts  
for a user — so that data can be accessible even when users are  
offline!

The role of `amphub` will also become less prominent during the  
development of Ampify 0.x. The introduction of the Amp Routing  
Protocol will allow for Host location and user Host updates to  
happen in a completely peer-to-peer manner.

This will leave `amphub` to simply act as a registry of "dumb  
numbers" to public keys. And various measures will be put in  
place to counter any denial-of-service attacks (on both legal  
and technical fronts).

In contrast to Ampify 0.x which will use IPv6 (using a Teredo-  
like service when native IPv6 isn't available), connections in  
Zero happens over IPv4.

Similarly, in contrast to using a UDP-based transport protocol  
with LEDBAT-like congestion control and SPDY-like framing  
with TLS-like crypto using OpenPGP-like certificates, Zero  
uses traditional HTTPS connections.

The location of a Host in Ampify Zero is defined as a set of  
either of the following pairs:

* FQDN (absolute domain name) + port
* IP address + port

It is expected that Hosts running on user devices will tend to  
use raw IP addresses, whereas Hosts running off servers will  
tend to use absolute domain names.

Since most "home users" will be behind NAT devices, Zero  
supports port mapping and NAT traversal using either UPnP  
IGD or NAT-PMP. If those don't work, it'd be up to the Host  
admins to manually configure their routers.

Ampify 0.x will have more comprehensive NAT traversal support  
and this will be tightly integrated with its use of tunneling  
when native IPv6 connectivity isn't available.

Ampify Zero frontends bind to an HTTPS server at port 8040  
by default. Host admins are expected to manually configure any  
LVS or similar proxies/load balancers to map port 443 to this  
port if they want to maximise connectivity.

Admins can take advantage of a number of both deployment and  
remote monitoring script hooks to simplify any such topology
related management.

Once a connection gets through, it is handled by nginx (the  
HTTPS frontend) which serves 3 purposes:

* Serve "root" static files, e.g. robots.txt, ampzero.js, etc.
* Proxy requests to `ampnode` instances
* Serve error pages, e.g. 50x when `amzero` instances are down

In contrast to Ampify 0.x, where requests will be dispatched to  
a request specific "nodule" app (which may even be compiled  
dynamically) on an appropriate internal Host node, the Ampify  
Zero design is super simple.

Nginx will proxy requests to an `ampnode` instance running on  
the same server. These instances will be homogeneous across a  
Host. That is, every single instance will be able to handle  
the exact same set of requests.

While this doesn't take advantage of locality in the way that 0.x  
would, it certainly makes for a simpler design. The `ampnode`  
instance is basically a combined event-driven and multi-threaded  
Python web server.

These instances are much like "app servers" and can be used in  
a manner similar to "modern" web app frameworks like Rails and  
Django. Hosts can define their own services to complement the  
built-in ones which all Ampnodes are expected to provide.

This is quite feeble in contrast to the intended LXC + Native  
Client sandboxed capability to run any arbitrary application code  
in Ampify 0.x, but it's also a lot simpler ;p

On startup, `ampnode` instances create a Redis sub-process for  
use as a shared memory store by various services. It also keeps  
in contact with a Keyspace cluster in order to partition and  
load balance the lexicographically-ordered "key space" of the  
Redis servers.

For example, imagine there are four `amzero` instances with a  
Redis server each. After a while, the respective "key space"  
they'd be responsible for might be split up like:

* redis-1, for all keys starting with A-F
* redis-2, for all keys starting with G-K
* redis-3, for all keys starting with L-S
* redis-4, for all keys starting with T-Z

Keyspace provides a Paxos-based lease mechanism and a nice  
strongly consistent datastore without a single-point-of-failure.  
The `ampnode` instances use it as a co-ordination space to  
manage 10-second leases of the "key space".

Note that, in contrast to the secure, asynchronous, Argonought  
encoded calls to other nodes in Ampify 0.x, Zero uses unencrypted  
TCP and process-specific protocols to talk to Redis/Keyspace.

It is up to Host admins to secure the network using OpenVPN  
or something similar if instances will be communicating over  
the public internet.

Some "state" is held/cached by Ampnodes, e.g. in Redis, as  
persistent connections, in the built-in filestorage service,  
etc. These are designed to be "revivable". That is, should the  
Ampnode die permanently, it should not cause any problems.

This is achieved by careful design of the built-in services and  
the use of durable stores, e.g. App Engine datastores for Ampify  
Items and S3 for files.

However, due to latency issues and limitations of the various  
durable stores, Ampnodes intelligently store copies "locally"  
in a manner suitable for Ampify applications.

Two custom App Engine applications have also been developed to  
help on this front. One is called `zerodata` — it provides a  
minimal wrapper around the App Engine datastore for storing  
Ampify Items and allows for both transactional writes and  
parallel queries to be executed.

The other is called `logstore`. This is intended to be used by  
the various Ampnodes as the place to log access, usage and  
errors. The advantage to writing to `logstore` instead of to local  
log files is two-fold:

* It provides a centralised location to see all the logs
* It allows for rich reports to be generated using mapreduce

If, for any reason, App Engine should be down, AmpNodes will  
try and provide as much of a working instance as possible, e.g.  
various bits of data will still be available in read-only mode  
and the instances will log to local files temporarily, etc.

The use of proprietary platforms like App Engine and S3 is only  
a temporary measure for Ampify Zero. The 0.x line will see the  
development of a scalable, strongly consistent, richly queryable,  
suited for data warehousing, live datastore called `ampstore`.

**Initialise A New Setup**

The `ampzero init` command is used to create a completely new  
instance. So, if you want to create an instance called `kickass`,  
you'd create it with:

    $ ampzero init kickass

The layout of the instance directory would look something like:

    kickass/
        .git/
        .gitignore
        README.md
        amprun.yaml
        ampnode/
            ampnode.yaml
            host-7.control.key
            host-7.private.key
            host-7.public.key
            hub.public.key
            redis.conf
            services/
            templates/
            user-23.control.key
            user-23.private.key
            user-23.public.key
        deployment.yaml
        keyspace/
            keyspace-single.conf
            keyspace-0.conf
            keyspace-1.conf
            keyspace-2.conf
        logstore/
            app.yaml
            index.yaml
            config.py
            lib/
            logstore.py
            remote.py
            www/
        nginx/
            nginx.conf
            server.cert
            www/
        zerodata/
            app.yaml
            index.yaml
            config.py
            lib/
            remote.py
            www/
            zerodata.py

It will be created as a sibling directory to wherever you'd checked out  
the  ampify repo. That is, if the ampify repo had been checked out into:

    /home/tav/repo/ampify

Then the new `kickass` instance will be at:

    /home/tav/repo/kickass

The are two reasons for this. First of all, the various commands tend towards  
a "[convention over configuration]" approach and many files are symlinked  
using relative paths.

Secondly, Ampify Zero is still in development and by using symlinks, it  
helps keep various files in sync without too much hassle.

Now when you run `ampzero init`, it will ask you various questions and will  
use your answers to create the relevant files — just hit enter to use the  
default values for the questions.

Most of the files can be easily re-created. And you can even over-write  
the files in an existing instance, e.g.

    $ ampzero init kickass --clobber

This will freshly re-create the various files as if you'd run init for the  
first time. It'd leave other files alone though. This is useful to keep  
up with any changes that might be available in the ampify repository.

However, 6 of the generated files are special and will not be clobbered.  
These contain public/private/control keys for you and your instance:

    ampnode/user-<id>.control.key
    ampnode/user-<id>.private.key
    ampnode/user-<id>.public.key
    ampnode/host-<id>.control.key
    ampnode/host-<id>.private.key
    ampnode/host-<id>.public.key

The public key components will have been signed by `amphub.org` and  
you'd have been given unique user and instance ID numbers when you  
first ran `ampzero init`.

The numbers in the `user-23.public.key` and `host-7.public.key`  
filenames in the instance directory listing above are references to the  
respective unique user and instance ID numbers.

You can share the public keys and ID numbers, but the private and control  
keys must be kept safe and private. The control keys in particular are used  
to update your keys with `amphub.org` — never share it!

And, finally, `ampzero init` will have automatically setup a git repository in  
the instance directory. You might want to push this to a private GitHub  
repo (or equivalent) for both backup and collaboration purposes.

The control keys will not be checked into this repo and are excluded via  
`.gitignore` — you should back them up separately and securely.

**Running Ampify Zero**

Once you have an instance setup, you can use `ampzero run` to run  
all the components at once, e.g. to run the above `kickass`, you'd:

    $ ampzero run kickass

Behind the scenes, this would start up a bunch of different processes:

* keyspace node (running in single mode)
* logstore server
* zerodata server
* redis servers
* ampnode instances
* nginx frontend

And, assuming that you'd chosen the default port settings, you  
should now be able to point your browser at the following URL and  
login:

* [https://localhost:8040]

The various log, pid and related files for these processes will be  
within the `amprun_root_directory` setting as specified in your  
`amprun.yaml` file and defaults to:

    /opt/ampzero/var/

When you run `ampzero init` it might run commands via sudo  
to fix the directory permissions. You can do this yourself by  
doing  something like:

    $ sudo mkdir /opt/ampzero
    $ sudo chown your-username /opt/ampzero

By default, the ampzero run process daemonises itself. You can  
suppress this with the `--no-daemon` parameter — allowing you  
to kill all the processes with a single ^C.

Otherwise, you can stop or force quit a running set of processes, e.g.

    $ ampzero run kickass stop
    $ ampzero run kickass quit

You can also enable debug mode settings, e.g.

    $ ampzero run kickass --debug

This also implicitly sets the `--no-daemon` parameter.

By default, your instance will communicate with `amphub.org`  
so that other instances can know how to contact you. You can  
disable this:

    $ ampzero run kickass --no-hub

However, until amp routing is developed, other nodes will not be  
able to contact you if your address changes and you're in this  
mode. It can, however, be quite useful whilst you're developing  
though.

And, on a related front, the `ampzero` run will try to figure  
out your public/external IP address and establish a port mapping  
with any NAT device that might be in between your machine and  
the internet. You can disable this, e.g.

    $ ampzero run kickass --no-nat-bypass

This setting is also implicit in the command that you'd use to  
run an instance on a "proper" server or VPS:

    $ ampzero run kickass --server

This mode differs from the default (which is suited for normal  
single-machine instances), in that it'd only start the following  
processes:

* redis servers
* ampnode instances
* nginx frontend

It assumes that you've got stable Keyspace servers running and  
that the zerodata/logstore apps are running — either on Google  
App Engine or elsewhere using [AppScale] or [TyphoonAE].

**Server/Cloud Deployment**

Until amp routing is developed, there's no built-in support to let  
other instances provide offline replication. So, in order to run an  
instance which is accessible when you're offline, you're going to  
have to deploy to a server with decent internet connectivity.

This is easy enough when you've got just a single server but  
becomes quite problematic once you have anything more than 2  
servers. So, to help with the problem, the `ampzero deploy`  
command tries to simplify things:

    $ ampzero deploy kickass

This will first generate tarballs of specific directories from your  
ampify and instance repositories. This step requires clean working  
directories, so either commit or `git stash` your changes first.

It will then communicate to the various `username@hosts/ips`  
that you've specified in your `deployment.yaml` file. This will  
happen over SSH, so you need to have the respective SSH keys.

You can limit deployment to specific hosts with:

    $ ampzero deploy kickass --host tav@212.4.3.2

You can also use an alternative file instead of `deployment.yaml`,  
e.g.

    $ ampzero deploy kickass --config ~/path/to/alt.yaml

The `ampzero deploy` will try to find out if there's an existing  
deployment, and if it's previously generated a tarball matching those  
versions, it'd  create a compressed diff to send — otherwise, it'd  
transfer over the full tarball.

And, finally, it'd run through a series of commands in order to switch  
over all of your servers in a synchronised manner:

* runs the `pre_deploy` script if you'd specified one
* builds new tarballs if remote architecture differs
* double-verifies tarballs/patches in case of memory corruption
* extracts and sets up the updated directories
* runs host-specific `pre_update` scripts if you'd specified any
* tells nginx to serve a placeholder "being upgraded" page
* runs `ampzero run stop` for any existing contexts
* switches over symlinks
* reloads nginx — upgrading to a new binary on the fly
* launches `ampzero run` in the new deployment
* tells nginx to go back to serving requests as normal
* cleans up all but the most recent of any previous deployments
* runs the `post_deploy` script if you'd specified one

The directory layout on your servers would end up looking something  
like:

    opt/
        ampzero/
            dist/
                .current
                .previous
                ampify/
                ampify-4f59ef0205/
                ampify-bf7663914a/
                instance/
                instance-5ec279ee02/
                instance-21e1bfc4f7/
                lock-file
                patches/
            var/

Now, the mechanism used by `ampzero deploy` is far from perfect.  
It is not robust against your servers failing and it wouldn't scale  
beyond 100 or so servers.

A far better solution, `redpill`, has already been designed and it'll  
be implemented as part of the Ampify development. It'll allow for  
scalable, failure-tolerant, synchronised updates and multi-version  
deployment.

But, for now, `ampzero deploy` is all we have.

**Customising Your Instance**

You can create custom services and modify the templates by  
changing the files inside your instance's `ampnode` directory.

Be sure to update the `ampnode.yaml` config file with the modules  
for any new services you define.

**Cheatsheet**

    $ ampzero init <instance-name>    # initialise a new setup
    $ ampzero run <instance-name>     # run all the components
    $ ampzero deploy <instance-name>  # deploy to remote hosts

**Resources**

You can use `--help` on all the above commands. For more info or  
to just say hi, come visit the:

* irc channel: [irc://irc.freenode.net/esp], [irc logs]
* mailing list: [http://groups.google.com/group/ampify]

**Contribute**

To contribute any patches, simply fork this repository using GitHub,  
add yourself to the `AUTHORS` file and send a pull request to me —  
[http://github.com/tav]. Thanks!


— Enjoy, tav <<tav@espians.com>>




[README]: http://github.com/tav/ampify/blob/master/README.md
[convention over configuration]: http://en.wikipedia.org/wiki/Convention_over_configuration
[https://localhost:8040]: https://localhost:8040/
[AppScale]: http://code.google.com/p/appscale/
[TyphoonAE]: http://code.google.com/p/typhoonae/

[http://github.com/tav]: http://github.com/tav
[http://ampify.it]: http://ampify.it
[http://ampify.it/planfile.html]: http://ampify.it/planfile.html
[http://groups.google.com/group/ampify]: http://groups.google.com/group/ampify
[irc://irc.freenode.net/esp]: irc://irc.freenode.net/esp
[irc logs]: http://irclogs.ampify.it
