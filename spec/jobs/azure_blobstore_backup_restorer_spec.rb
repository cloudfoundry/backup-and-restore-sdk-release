require 'rspec'
require 'yaml'
require 'bosh/template/test'
require 'json'

describe 'azure-blobstore-backup-restorer job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../..')) }
  let(:job) { release.job('azure-blobstore-backup-restorer') }
  let(:backup_template) { job.template('bin/bbr/backup') }
  let(:restore_template) { job.template('bin/bbr/restore') }
  let(:containers_template) { job.template('config/containers.json') }

  describe 'backup' do
    context 'when backup is not enabled' do
      it 'the templated script is empty' do
        config = backup_template.render({})
        expect(config.strip).to eq("#!/usr/bin/env bash\n\nset -eu")
      end

      it 'the bucket config is empty' do
        manifest = {
          "enabled" => false,
          "containers" => {
            "droplets" => nil
          }
        }
        config = containers_template.render(manifest)
        expect(config.strip).to eq("")
      end
    end

    context 'when backup is enabled' do
      context 'and bpm is enabled' do
        it 'templates bpm command correctly' do
          config = backup_template.render({"bpm" => {"enabled" => true}, "enabled" => true})
          expect(config).to include("/var/vcap/jobs/bpm/bin/bpm run azure-blobstore-backup-restorer")
        end
      end

      context 'and bpm is not enabled' do
        it 'does not template bpm' do
          config = backup_template.render("enabled" => true)
          expect(config).to include("backup")
          expect(config).not_to include("/var/vcap/jobs/bpm/bin/bpm run azure-blobstore-backup-restorer")
        end
      end
    end
  end

  describe 'restore' do
    context 'when restore is not enabled' do
      it 'the templated script is empty' do
        config = restore_template.render({})
        expect(config.strip).to eq("#!/usr/bin/env bash\n\nset -eu")
      end
    end

    context 'when restore is enabled' do
      context 'and when bpm is enabled' do
        it 'templates bpm command correctly' do
          config = restore_template.render({"bpm" => {"enabled" => true}, "enabled" => true})
          expect(config).to include("/var/vcap/jobs/bpm/bin/bpm run azure-blobstore-backup-restorer")
        end
      end

      context 'when bpm is not enabled' do
        it 'does not template bpm' do
          config = restore_template.render("enabled" => true)
          expect(config).to include("restore")
          expect(config).not_to include("/var/vcap/jobs/bpm/bin/bpm run azure-blobstore-backup-restorer")
        end
      end
    end
  end
end
