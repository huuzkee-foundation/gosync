#!/bin/bash

sudo yum install -y mysql-server
sudo service mysqld start
mysql -u root -e "create database gosync; grant all privileges on gosync.* to 'test'@'%' identified by 'testing' with grant option; flush privileges;"
