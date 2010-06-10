**Cheatsheet**

See the sections below for further info on each command:

    $ ampinit <instance-name>    # initialise a new setup
    $ amprun <instance-name>     # run all the components
    $ ampdeploy <instance-name>  # deploy to remote hosts

**Introduction**

Assuming that you've successfully built all the dependencies according  
to the root [README] file, you should now be all set to run Ampify Zero.

Unfortunately, there are a number of different components involved in an  
Ampify Zero setup (Keyspace cluster, Redis servers, Google App Engine  
instances and ampzero itself).

And unless you're used to modern "high-scalability" deployments in other  
contexts, it could all be quite a head fuck. Therefore, a few utility scripts  
are provided to help ease the pain.

**Initialise A New Setup**

The `ampinit` script is used to create a completely new instance. So, if  
you want to create an instance called `kickass`, you'd create it with:

    $ ampinit kickass

The layout of the instance directory would look like:

    kickass/
        .git/
        README.md
        amprun.yaml
        ampzero/
            admin-user.info
            ampzero.yaml
            hub.info
            provider.info
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
            logstore.py
            lib/
            www/
        nginx/
            nginx.conf
            server.cert
            www/
        zerodata/
            app.yaml
            index.yaml
            config.py
            zerodata.py
            lib/
            www/

It will be created as a sibling directory to wherever you'd checked out  
the  ampify repo. That is, if the ampify repo had been checked out into:

    /home/tav/repo/ampify

Then the new `kickass` instance will be at:

    /home/tav/repo/kickass

The are two reasons for this. First of all, the various scripts tend towards  
a "[convention over configuration]" approach and many files are symlinked  
using relative paths.

Secondly, Ampify Zero is still in development and by using symlinks, it  
helps keep various files in sync without too much hassle.

Now when you run `ampinit`, it will ask you various questions and will  
use your answers to create the relevant files — just hit enter to use the  
default values for the questions.

Most of the files can be easily re-created. And you can even over-write  
the files in an existing instance, e.g.

    $ ampinit kickass --clobber

This is useful to keep up with any changes that might be available in the  
ampify repository.

However, 2 of the generated files are special and will not be clobbered.  
These contain public/private key-pairs for you and your instance:

    ampzero/admin-user.info
    ampzero/provider.info

The public key components will have been signed by `amphub.org` and  
you'd have been given unique user and instance ID numbers when you  
first ran `ampinit`.

You can share the public key and IDs, but the private key bits must be  
kept safe and private — you might even want to back them up... =)

And, finally, `ampinit` will have automatically setup a git repository in  
the instance directory. You might want to push this to a private GitHub  
repo (or equivalent) for both backup and collaboration purposes.

**Running Ampify Zero**

Once you have an instance setup, you can use `amprun` to run all the  
components at once, e.g. to run the above `kickass`, you'd:

    $ amprun kickass

Behind the scenes, this would start up a bunch of different processes:

* keyspace node (running in single mode)
* logstore server
* zerodata server
* redis servers
* ampzero instances
* nginx frontend

And, assuming that you'd chosen the default port settings, you should  
now be able to point your browser at the following URL and login:

* [https://localhost:8040]

The various log, pid and related files for these processes will be within  
the `amprun_root` setting as specified in your `amprun.yaml` file  
and defaults to:

    /opt/ampzero/var/

You might want to fix up the directory permissions by doing something  
like:

    $ sudo mkdir /opt/ampzero
    $ sudo chown your-username /opt/ampzero

By default, the amprun process daemonises itself. You can suppress  
this with the `--no-daemon` parameter — allowing you to kill all the  
processes with a single ^C.

Otherwise, you can stop or force quit a running set of processes, e.g.

    $ amprun kickass stop
    $ amprun kickass quit

You can also enable debug mode settings, e.g.

    $ amprun kickass --debug

This also implicitly sets the `--no-daemon` parameter.

By default, your instance will communicate with `amphub.org` so that  
other instances can know how to contact you. You can disable this:

    $ amprun kickass --no-hub

However, until amp routing is developed, other nodes will not be able  
to contact you if your address changes and you're in this mode. It can,  
however, be quite useful whilst you're developing though.

And, on a related front, amprun will try to figure out your public/external  
IP address and establish a port mapping with any NAT device that might  
be in between your machine and the internet. You can disable this, e.g.

    $ amprun kickass --no-nat-bypass

This setting is also implicit in the command that you'd use to run an  
instance on a "proper" server or VPS:

    $ amprun kickass --server

This mode differs from the default (which is suited for single-machine  
instances), in that it'd only start the following processes:

* redis servers
* ampzero instances
* nginx frontend

It assumes that you've got stable Keyspace servers running and that  
the zerodata/logstore apps are running — either on Google App Engine  
or elsewhere using [TyphoonAE].

**Server/Cloud Deployment**

Until amp routing is developed, there's no built-in support to let other  
instances provide offline replication. So, in order to run an instance  
which is accessible when you're offline, you're going to have to deploy  
to a server with decent internet connectivity.

This is easy enough when you've got just a single server but becomes  
quite problematic once you have anything more than 2 servers. So, to  
help with the problem, the `ampdeploy` script tries to simplify things:

    $ ampdeploy kickass

This will first generate tarballs of specific directories from your ampify  
and instance repositories. This step requires clean working directories,  
so either commit or `git stash` your changes first.

The script will then communicate to the various `username@hosts`  
that you've specified in your `deployment.yaml` file. This will happen  
over SSH, so you need to have SSH servers running and the keys.

You can limit deployment to specific hosts with:

    $ ampdeploy kickass --host tav@212.4.3.2

You can also use an alternative file instead of `deployment.yaml`,  
e.g.

    $ ampdeloy kickass --config ~/path/to/alt.yaml

The script will try to find out if there's an existing deployment, and  
if it's previously generated a tarball matching those versions, it'd  create  
a compressed diff to send — otherwise, it'd transfer over the full tarball.

And, finally, it'd run through a series of commands in order to switch  
over all of your servers in a synchronised manner:

* double-verify tarballs/patches in case of memory corruption
* extract and setup the updated directories
* run `amprun stop` for any existing contexts
* switch over symlinks
* launch `amprun` in the new deployment
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

Now, the mechanism used by `ampdeploy` is far from perfect. It is not  
robust against your servers failing and it wouldn't scale beyond 100 or so  
servers.

A far better solution, `redpill`, has already been designed and it'll be  
implemented as part of the Ampify development. It'll allow for scalable,  
failure-tolerant, synchronised updates and multi-version deployment.

But, for now, `ampdeploy` is all we have.

**Customising Your Instance**

You can create custom services and modify the templates by changing  
the files inside your instance's ``ampzero`` directory.

Be sure to update the `ampzero.yaml` config file with the modules for  
any new services you define.


**Resources**

You can use `--help` on all the above scripts. For more info or even to  
just say hi, visit the:

* irc channel: [irc://irc.freenode.net/esp], [irc logs]
* mailing list: [http://groups.google.com/group/ampify]
* online docs: [http://ampify.it]

**Contribute**

To contribute any patches, simply fork this repository using GitHub,  
add yourself to the `AUTHORS` file and send a pull request to me —  
[http://github.com/tav]. Thanks!


— Enjoy, tav <<tav@espians.com>>




[README]: ../../README.md
[convention over configuration]: http://en.wikipedia.org/wiki/Convention_over_configuration
[https://localhost:8040]: https://localhost:8040/
[TyphoonAE]: http://code.google.com/p/typhoonae/

[http://github.com/tav]: http://github.com/tav
[http://ampify.it]: http://ampify.it
[http://ampify.it/planfile.html]: http://ampify.it/planfile.html
[http://groups.google.com/group/ampify]: http://groups.google.com/group/ampify
[irc://irc.freenode.net/esp]: irc://irc.freenode.net/esp
[irc logs]: http://irclogs.ampify.it
