Vagrant.configure("2") do |config|
  config.ssh.forward_x11 = true
  config.ssh.forward_agent = true

  config.vm.provider "virtualbox" do |vb|
    vb.customize ["modifyvm", :id, "--memory", "1024"]
  end

  config.vm.box = "ubuntu/trusty64"
  config.vm.hostname = "armillaria"

  config.ssh.insert_key = false
  config.vm.network "private_network", ip: "192.168.50.13"

  # Temporary fix in order to make docker install on ubuntu/thrusty64,
  # until this issue is resolved:
  # https://github.com/mitchellh/vagrant/issues/5697
  config.vm.provision :shell,
    inline: "sudo apt-get update"

  config.vm.provision :docker do |d|

    #https://github.com/tenforce/docker-virtuoso
    d.run "virtuoso",
      image: "tenforce/virtuoso",
      cmd: "/bin/bash /startup.sh",
      args: "-p 8890:8890 -p 1111:1111 \
            -e DBA_PASSWORD=secret \
            -e SPARQL_UPDATE=true"
  
    #https://github.com/dockerfile/elasticsearch
    d.run "elasticsearch",
      image: "elasticsearch:latest",
        args: "-p 9200:9200 -p 9300:9300"

    #https://github.com/digibib/armillaria
    d.build_image "/vagrant",
      args: "-t armillaria"
  end

  # Allow Elasticsearch and Virtuoso dockers 3 secs to spin up before Armillaria
  config.vm.provision :shell,
    inline: "sleep 3"

  config.vm.provision :docker do |d|
    d.run "armillaria",
      args: "-p 8080:8080 \
            -e SERVER_PORT=8080 \
            --link elasticsearch:elasticsearch \
            --link virtuoso:virtuoso"
  end
end
