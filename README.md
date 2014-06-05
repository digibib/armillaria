# Armillaira
RDF cataloging interface.

## Setup
Eventually, a debian package will be provided, but for now, you must manually set up your environment.

The instructions will probably work with any Debian/Ubuntu-flavoured distro. My development setup is Ubuntu 64-bit version 14.04.

### Runtime dependencies

#### Openlink Virtuoso quad-store
Armillaria is developed and tested against Virtuoso 7, which must be compiled [from source](https://github.com/openlink/virtuoso-opensource):

```bash
TODO build instructions
```


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

There are pacakges for Ubuntu:

```bash
sudo apt-get install golang
```

If you want the latest and greatest I recommend the [official binary distributions](https://code.google.com/p/go/wiki/Downloads?tm=2):

```bash
wget https://storage.googleapis.com/golang/go1.2.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.2.2.linux-amd64.tar.gz
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
make indexes
```

To start the server type:

```bash
make run
```
