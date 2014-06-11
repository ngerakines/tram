require 'chefspec'
require 'chefspec/berkshelf'
ChefSpec::Coverage.start!

platforms = {
  'ubuntu' => ['12.04', '13.10'],
  'centos' => ['5.9', '6.5']
}

describe 'tram::app' do

  platforms.each do |platform_name, platform_versions|

    platform_versions.each do |platform_version|

      context "no install type on #{platform_name} #{platform_version}" do

        let(:chef_run) do
          ChefSpec::Runner.new(platform: platform_name, version: platform_version) do |node|
            node.set['preview']['install_type'] = 'none'
          end.converge('tram::app')
        end

        it 'includes dependent receipes' do
          expect(chef_run).to include_recipe('apt')
          expect(chef_run).to include_recipe('yum')
        end

        it 'creates the user and groups' do
          expect(chef_run).to create_user('tram')
          expect(chef_run).to create_group('tram')
        end

        it 'installs required packages' do
          expect(chef_run).to install_package('unzip')
        end

      end

      context "package install type on #{platform_name} #{platform_version}" do

        let(:chef_run) do
          ChefSpec::Runner.new(platform: platform_name, version: platform_version) do |node|
            node.set['tram']['install_type'] = 'package'
          end.converge('tram::app')
        end

        it 'installs the tram package' do
          expect(chef_run).to install_package('tram')
        end

      end

      context "archive install type on #{platform_name} #{platform_version}" do

        let(:chef_run) do
          ChefSpec::Runner.new(platform: platform_name, version: platform_version) do |node|
            node.set['tram']['install_type'] = 'archive'
          end.converge('tram::app')
        end

        it 'installs the tram archive and unpacks it' do
          expect(chef_run).to create_remote_file('/var/chef/cache/tram.zip')
          expect(chef_run).to run_bash('extract_app')
          expect(chef_run).to run_execute('chown -R tram:tram /home/tram/')
          expect(chef_run).to create_file('/home/tram/tram')
        end

      end

    end

  end

end
