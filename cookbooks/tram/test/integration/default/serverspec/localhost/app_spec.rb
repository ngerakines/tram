require 'spec_helper'

describe 'tram package' do

  describe user('tram') do
    it { should exist }
  end

  describe group('tram') do
    it { should exist }
  end

  describe file('/home/tram/tram') do
    it { should be_file }
    it { should be_owned_by 'tram' }
    it { should be_grouped_into 'tram' }
    it { should be_executable }
  end

  describe file('/etc/init.d/tram') do
    it { should be_file }
    it { should be_owned_by 'root' }
    it { should be_grouped_into 'root' }
  end

  describe port(7040) do
    it { should be_listening }
  end

end
