#!/bin/bash

sudo yum install -y wget
wget http://mirrors.nl.eu.kernel.org/fedora-epel/6/i386/epel-release-6-8.noarch.rpm
sudo rpm -Uvh epel-release-6-8.noarch.rpm
rm -f epel-release-6-8.noarch.rpm
sudo yum install -y golang
export GOPATH=$GOPATH:/data/go:/data/go/src/github.com/Nitecon/gosync/Godeps/_workspace
if [ ! -f /etc/profile.d/gosync.sh ]; then
    echo 'export GOPATH=$GOPATH:/data/go:/data/go/src/github.com/Nitecon/gosync/Godeps/_workspace' > gosync_profile.sh
    echo 'alias gosyncDir="cd /data/go/src/github.com/Nitecon/gosync"' >> gosync_profile.sh
    sudo mv gosync_profile.sh /etc/profile.d/gosync.sh
    echo "Alias added: gosyncDir" > motd
    echo "-- The alias will quickly get you to the working directory for gosync" >> motd
    echo "-- Usage: Just type gosyncDir anywhere and hit Enter Key" >> motd
    sudo mv -f motd /etc/motd
fi

sudo mkdir /data/storage

#TODO: Add crons to modify /data/storage to validate file changes

