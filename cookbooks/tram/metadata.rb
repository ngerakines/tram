name             'tram'
maintainer       'Nick Gerakines'
maintainer_email 'nick@gerakines.net'
license          'MIT'
description      'Installs/Configures tram'
long_description IO.read(File.join(File.dirname(__FILE__), 'README.md'))
version          '1.0.0'

depends 'apt'
depends 'yum'

supports 'centos'
supports 'ubuntu'
