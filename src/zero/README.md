
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
instances and ampzero itself).

And unless you're used to modern "high-scalability" deployments in other  
contexts, it could all be quite a head fuck. Therefore, a utility script  
is provided to help ease the pain.

**Initialise A New Setup**

The `ampzero init` command is used to create a completely new instance. So, if  
you want to create an instance called `kickass`, you'd create it with:

    $ ampzero init kickass

The layout of the instance directory would look like:

    kickass/
        .git/
        .gitignore
        README.md
        amprun.yaml
        ampzero/
            admin-user.control.key
            admin-user.private.key
            admin-user.public.key
            ampzero.yaml
            host.control.key
            host.private.key
            host.public.key
            hub.public.key
            redis.conf
            services/
            templates/
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

This is useful to keep up with any changes that might be available in the  
ampify repository.

However, 6 of the generated files are special and will not be clobbered.  
These contain public/private/control keys for you and your instance:

    ampzero/admin-user.control.key
    ampzero/admin-user.private.key
    ampzero/admin-user.public.key
    ampzero/host.control.key
    ampzero/host.private.key
    ampzero/host.public.key

The public key components will have been signed by `amphub.org` and  
you'd have been given unique user and instance ID numbers when you  
first ran `ampzero init`.

You can share the public key and ID numbers, but the private and control  
keys must be kept safe and private. The control key in particular is used  
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
* ampzero instances
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

And, on a related front, ampzero run will try to figure out your  
public/external IP address and establish a port mapping with  
any NAT device that might be in between your machine and the  
internet. You can disable this, e.g.

    $ ampzero run kickass --no-nat-bypass

This setting is also implicit in the command that you'd use to  
run an instance on a "proper" server or VPS:

    $ ampzero run kickass --server

This mode differs from the default (which is suited for normal  
single-machine instances), in that it'd only start the following  
processes:

* redis servers
* ampzero instances
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

* build new tarballs if remote architecture differs
* double-verify tarballs/patches in case of memory corruption
* extract and setup the updated directories
* run `ampzero run stop` for any existing contexts
* switch over symlinks
* launch `ampzero run` in the new deployment
* clean up all but the most recent of any previous deployments

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
changing the files inside your instance's ``ampzero`` directory.

Be sure to update the `ampzero.yaml` config file with the modules  
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
