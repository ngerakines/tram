#
# Cookbook Name:: tram
# Recipe:: app
#
# Copyright (C) 2014 Nick Gerakines <nick@gerakines.net>
# 
# This project and its contents are open source under the MIT license.
#

node.set["monit"]["reload_on_change"] = false

include_recipe 'apt'
include_recipe 'yum'

if node[:tram][:enable_monit] then
  include_recipe 'monit::default'
end

if node[:tram][:enable_logrotate] then
  include_recipe 'logrotate::default'
end

require 'json'

user 'tram' do
  username 'tram'
  home '/home/tram'
  action :remove
  action :create
  supports ({ :manage_home => true })
end

group 'tram' do
  group_name 'tram'
  members 'tram'
  action :remove
  action :create
end

template '/etc/tram.conf' do
  source 'tram.conf.erb'
  mode 0640
  group 'tram'
  owner 'tram'
  variables(:json => JSON.pretty_generate(node[:tram][:config].to_hash))
end

case node[:tram][:install_type]
when 'package'

  package node[:tram][:package]

when 'archive'

  %w{unzip}.each do |pkg|
    package pkg
  end

  remote_file "#{Chef::Config[:file_cache_path]}/tram.zip" do
    source node[:tram][:package_source]
  end

  bash 'extract_app' do
    cwd '/home/tram/'
    code <<-EOH
      unzip #{Chef::Config[:file_cache_path]}/tram.zip
      EOH
    not_if { ::File.exists?('/home/tram/tram') }
  end

  execute 'chown -R tram:tram /home/tram/'

  file '/home/tram/tram' do
    mode 00777
  end

end

cookbook_file '/etc/init.d/tram' do
  source 'tram'
  mode 00777
  owner 'root'
  group 'root'
end

service 'tram' do
  provider Chef::Provider::Service::Init
  action [:start]
end

if node[:tram][:enable_monit] then
  monit_monitrc 'tram' do
    variables({ category: 'tram' })
  end
end

if node[:tram][:enable_logrotate] then
  logrotate_app 'tram' do
    cookbook  'logrotate'
    path      ['/var/log/tram.log']
    options   ['missingok', 'delaycompress', 'copytruncate']
    frequency 'daily'
    size      1048576
    maxsize   2097152
    rotate    2
    create    '644 root root'
  end
end
