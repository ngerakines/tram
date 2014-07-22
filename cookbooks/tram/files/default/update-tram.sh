#!/bin/sh

chef-client

if [ -f /home/tram/tram ]
then
	/etc/init.d/monit stop
	/etc/init.d/tram stop
	yum clean all
	yum update -y tram
	/etc/init.d/tram start
	/etc/init.d/monit start
fi
