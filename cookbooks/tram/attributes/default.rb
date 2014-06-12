
default[:tram][:platform] = 'amd64'

default[:tram][:version] = '1.0.0'
default[:tram][:install_type] = 'archive'
default[:tram][:package] = 'tram'
default[:tram][:package_source] = "https://github.com/ngerakines/tram/releases/download/v#{node[:tram][:version]}/tram-#{node[:tram][:version]}.linux_#{node[:tram][:platform]}.zip"

default[:tram][:port] = 7040
default[:tram][:basePath] = "/home/tram/data"

default[:tram][:config] = {}
default[:tram][:config][:listen] = ":#{node[:tram][:port]}"
default[:tram][:config][:lruSize] = 120000

default[:tram][:config][:index] = {}
default[:tram][:config][:index][:engine] = "local"
default[:tram][:config][:index][:localBasePath] = "#{node[:tram][:basePath]}/index/"

default[:tram][:config][:storage] = {}
default[:tram][:config][:storage][:engine] = "local"
default[:tram][:config][:storage][:localBasePath] = "#{node[:tram][:basePath]}/assets/"
