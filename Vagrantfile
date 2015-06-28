Vagrant.configure("2") do |config|
    config.vm.box = "centos"
    config.vm.box_url = "https://github.com/2creatives/vagrant-centos/releases/download/v6.5.3/centos65-x86_64-20140116.box"
    config.vm.synced_folder "./", "/data/go/src/github.com/Nitecon/gosync", create:true

    config.vm.define "syncnode1" do |syncnode1|
        syncnode1.vm.network :private_network, ip: "10.0.1.100"
        syncnode1.vm.provision :shell, :path => "vagrant_tests/vagrant_box_node.sh"
        syncnode1.vm.box = "centos"
    end

    config.vm.define "syncnode2" do |syncnode2|
        syncnode2.vm.network :private_network, ip: "10.0.1.101"
        syncnode2.vm.provision :shell, :path => "vagrant_tests/vagrant_box_node.sh"
        syncnode2.vm.box = "centos"
    end

    config.vm.define "db" do |db|
        db.vm.network :private_network, ip: "10.0.1.105"
        db.vm.provision :shell, :path => "vagrant_tests/vagrant_box_db.sh"
        db.vm.box = "centos"
    end
end
