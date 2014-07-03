# Armillaira
RDF cataloging interface.

## Architecture
[A coarse overview of the system architecture](https://docs.google.com/drawings/d/1t1vchNhGd86gPENq5_H0lCoWZGQYu8LqWr56qcUDbAU)

## Setup
Eventually, a debian package will be provided, but for now, you must manually set up your environment.

The instructions will probably work with any Debian/Ubuntu-flavoured distro. My development setup is Ubuntu 64-bit version 14.04.

### Runtime dependencies

#### Openlink Virtuoso quad-store
Ideally, any RDF quad-store with full SPARQL 1.1 support should do the trick, but Armillaria is developed and tested against Virtuoso 7, so that is what we recommend.

There are no binary distribution of Virtuoso 7, so it which must be compiled [from source](https://github.com/openlink/virtuoso-opensource):

```bash
# fetch build dependencies
sudo apt-get install libssl-dev gawk gperf flex build-essential automake autoconf bison libtool
# fetch virtuoso source tree
git clone https://github.com/openlink/virtuoso-opensource.git
git checkout develop/7
# bootstrap build
CFLAGS="-O2 -m64" ./autogen.sh
# build (substitute your destination)
CFLAGS="-O2 -m64" ./configure --prefix=/your/install/destination --disable-dbpedia-vad --disable-demo-vad --disable-fct-vad --disable-isparql-vad --disable-ods-vad --disable-rdfmappers-vad --disable-rdb2rdf-vad --disable-sparqldemo-vad --disable-syncml-vad --disable-tutorial-vad --disable-bpel-vad --with-port=1111
```

Test your installation by running virtuoso in the foreground. You need to give it a configuration file like this:

```bash
/your/install/destination/bin/virtuoso-t -f -c /my/db/destination/virtuoso.ini
```

You may use [our minimal virtuoso.ini](https://github.com/digibib/saltdeploy/blob/master/koha/salt/koha/files/virtuoso/virtuoso.ini.minimal) as a starting point.

For a more robust deployment, you may want to run virtuoso as an upstart service: [example virtuoso.conf](https://github.com/digibib/saltdeploy/blob/master/koha/salt/koha/files/virtuoso/virtuoso.conf)


#### Elasticsearch
I recommend [installing a package from the official repostitories](http://www.elasticsearch.org/guide/en/elasticsearch/reference/current/setup-repositories.html):

```bash
wget -O - http://packages.elasticsearch.org/GPG-KEY-elasticsearch | sudo apt-key add -
echo "deb http://packages.elasticsearch.org/elasticsearch/1.1/debian stable main" | sudo tee -a /etc/apt/sources.list
sudo apt-get update && sudo apt-get install elasticsearch
sudo service elasticsearch start
```

### Development dependencies

#### Go
The (very limited) server backend is coded in Go.

There are pacakges for Ubuntu which are usually fairly up to date if you are running an updated distro:

```bash
sudo apt-get install golang
```

The minimum required version is version 1.2. If you want the latest and greatest I recommend  the [official binary distributions](https://code.google.com/p/go/wiki/Downloads?tm=2):

```bash
wget https://storage.googleapis.com/golang/go1.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

Either way, you must make sure to set your `GOPATH` environment variable. This tells Go where to install dependencies locally. After installing Go, type `go env` to check if `GOPATH` is set, otherwise do it yourself:

```bash
export GOPATH=/home/<your username>/go  # for example
```

Of course, you should add the above `export`-statements to your `.bashrc`, `.profile` or similar, to avoid having to enter it every time you open a new shell.

If you want to know more about using the Go tools, please consult its [exellent documentation](http://golang.org/doc/code.html).

### Configuration
Armellaria has a configuration file where you must fill out a few things:

```bash
mv data/config_example.json data/config.json
editor data/config.json
```

Fetch dependencies and setup indexes (expects Elasticsearch to be running on localhost:9200):

```bash
make deps
make indexes && make mappings
```

To start the server type:

```bash
make run
```
