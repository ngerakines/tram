name             'tram'
maintainer       'Nick Gerakines'
maintainer_email 'nick@gerakines.net'
license          'MIT'
description      'Installs/Configures tram'
long_description IO.read(File.join(File.dirname(__FILE__), 'README.md'))
version          '1.1.0'

depends 'apt'
depends 'yum'
depends 'yum-epel'
depends 'monit', '~> 1.5.3'
depends 'logrotate', '~> 1.5.0'
depends 'build-essential', '~> 2.0'

supports 'centos'
supports 'ubuntu'
