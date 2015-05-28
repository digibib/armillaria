# Bug in docker/packer needs to append slash in salt paths:
# https://github.com/mitchellh/packer/issues/1040
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

  config.vm.provision :shell, inline: <<SCRIPT
  apt-get install -y golang
SCRIPT

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

    d.run "armillaria",
      args: "-p 8080:8080 \
            -e SERVER_PORT=8080"
  end

end
