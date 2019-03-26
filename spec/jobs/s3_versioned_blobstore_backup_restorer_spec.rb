require 'rspec'
require 'yaml'
require 'bosh/template/test'
require 'json'

describe 's3-versioned-blobstore-backup-restorer job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../..')) }
  let(:job) { release.job('s3-versioned-blobstore-backup-restorer') }
  let(:buckets_template) { job.template('config/buckets.json') }
  let(:backup_template) { job.template('bin/bbr/backup') }
  let(:restore_template) { job.template('bin/bbr/restore') }
  let(:metadata_template) { job.template('bin/bbr/metadata') }

  describe 'backup' do
    context 'when backup is not enabled' do
      it 'the templated script is empty' do
        config = backup_template.render({})
        expect(config.strip).to eq("#!/usr/bin/env bash\n\nset -eu")
      end

      it 'the bucket config is empty' do
        manifest = {
          "enabled" => false,
          "buckets" => {
            "droplets"  => nil
            }
          }
        config = buckets_template.render(manifest)
        expect(config.strip).to eq("")
      end

      it 'the metadata script enables the skip_bbr_scripts flag' do
        metadata = metadata_template.render({})
        expect(metadata).to include("skip_bbr_scripts: true")
      end
    end

    context 'when backup is enabled' do
      it 'the metadata script disables the skip_bbr_scripts flag' do
        metadata = metadata_template.render("enabled" => true)
        expect(metadata).to include("skip_bbr_scripts: false")
      end

      context 'and bpm is enabled' do
        it 'templates bpm command correctly' do
          config = backup_template.render({"bpm" => {"enabled" => true}, "enabled" => true})
          expect(config).to include("/var/vcap/jobs/bpm/bin/bpm run s3-versioned-blobstore-backup-restorer")
        end
      end

      context 'and bpm is not enabled' do
        it 'does not template bpm' do
          config = backup_template.render("enabled" => true)
          expect(config).to include("backup")
          expect(config).not_to include("/var/vcap/jobs/bpm/bin/bpm run s3-versioned-blobstore-backup-restorer")
        end
      end

      context 'and it is configured correctly' do
        it 'succeeds' do
          manifest = {
            "enabled" => true,
            "buckets" => {
              "droplets"  => {
                "name" => "the_droplets_bucket",
                "region" => "eu-west-1",
                "aws_access_key_id" => "AWS_ACCESS_KEY_ID",
                "aws_secret_access_key" => "AWS_SECRET_ACCESS_KEY",
                "endpoint" => "endpoint_to_s3_compatible_blobstore",
                "use_iam_profile" => false,
                "backup" => {
                  "name" => "the_backup_droplets_bucket",
                  "region" => "eu-west-2",
                }
              }
            }
          }
          expect { backup_template.render(manifest) }.to_not(raise_error)
          expect { buckets_template.render(manifest) }.to_not(raise_error)
        end
      end

      context 'both secrets keys and an IAM profile are configured' do
        it 'errors' do
          manifest = {
            "enabled" => true,
            "buckets" => {
              "droplets"  => {
                "name" => "the_droplets_bucket",
                "region" => "eu-west-1",
                "aws_access_key_id" => "AWS_ACCESS_KEY_ID",
                "aws_secret_access_key" => "AWS_SECRET_ACCESS_KEY",
                "endpoint" => "endpoint_to_s3_compatible_blobstore",
                "use_iam_profile" => true,
                "backup" => {
                  "name" => "the_droplets_backup_bucket",
                  "region" => "eu-west-2",
                }
              },
            }
          }

         expect { buckets_template.render(manifest) }.to(raise_error(RuntimeError, 'Invalid configuration, both the access key ID and the secret key pair and an IAM profile were used for bucket droplets'))
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
          expect(config).to include("/var/vcap/jobs/bpm/bin/bpm run s3-versioned-blobstore-backup-restorer")
        end
      end

      context 'when bpm is not enabled' do
        it 'does not template bpm' do
          config = restore_template.render("enabled" => true)
          expect(config).to include("restore")
          expect(config).not_to include("/var/vcap/jobs/bpm/bin/bpm run s3-versioned-blobstore-backup-restorer")
        end
      end
    end
  end
end
