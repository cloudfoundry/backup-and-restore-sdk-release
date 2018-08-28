require 'rspec'
require 'yaml'
require 'bosh/template/test'

describe 'azure-blobstore-backup-restorer job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../..')) }
  let(:job) { release.job('azure-blobstore-backup-restorer') }
  let(:template) { job.template('bin/bbr/backup') }

  describe 'backup' do
    context 'when backup is not enabled' do
      it 'the templated script is empty' do
        config = template.render({})
        expect(config).to eq("#!/usr/bin/env bash\n\nset -eu\n\n")
      end
    end

    context 'when backup is enabled' do
      context 'and bpm is enabled' do
        it 'templates bpm command correctly' do
          config = template.render({"bpm" => {"enabled" => true}, "enabled" => true})
          expect(config).to include("/var/vcap/jobs/bpm/bin/bpm run azure-blobstore-backup-restorer")
        end
      end

      context 'and bpm is not enabled' do
        it 'does not template bpm' do
          config = template.render("enabled" => true)
          expect(config).not_to include("/var/vcap/jobs/bpm/bin/bpm run azure-blobstore-backup-restorer")
        end
      end
    end
  end
end