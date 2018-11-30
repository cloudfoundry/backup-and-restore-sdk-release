require 'rspec'
require 'yaml'
require 'bosh/template/test'

describe 's3-unversioned-blobstore-backup-restorer job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../..')) }
  let(:job) { release.job('s3-unversioned-blobstore-backup-restorer') }
  let(:backup_template) { job.template('bin/bbr/backup') }
  let(:buckets_template) { job.template('config/buckets.json') }
  let(:restore_template) { job.template('bin/bbr/restore') }

  describe 'backup' do
    context 'when backup is not enabled' do
      it 'the templated script is empty' do
        config = backup_template.render({})
        expect(config.strip).to eq("#!/usr/bin/env bash\n\nset -eu")
      end
    end

    context 'when backup is enabled' do
      context 'and bpm is enabled' do
        it 'templates bpm command correctly' do
          config = backup_template.render({"bpm" => {"enabled" => true}, "enabled" => true})
          expect(config).to include("/var/vcap/jobs/bpm/bin/bpm run s3-unversioned-blobstore-backup-restorer")
        end
      end

      context 'and bpm is not enabled' do
        it 'does not template bpm' do
          config = backup_template.render("enabled" => true)
          expect(config).to include("backup")
          expect(config).not_to include("/var/vcap/jobs/bpm/bin/bpm run s3-unversioned-blobstore-backup-restorer")
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

      context 'and the backup bucket is the same as the live bucket' do
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
                "use_iam_profile" => false,
                "backup" => {
                  "name" => "the_droplets_bucket",
                  "region" => "eu-west-2",
                }
              }
            }
          }
          expect { buckets_template.render(manifest) }.to(raise_error(RuntimeError, 'Invalid bucket configuration for droplets, name and backup.name must be distinct'))
        end
      end

      context 'and the backup bucket is a different live bucket' do
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
                "use_iam_profile" => false,
                "backup" => {
                  "name" => "my_packages_bucket",
                  "region" => "eu-west-2",
                }
              },
              "packages"  => {
                "name" => "my_packages_bucket",
                "region" => "eu-west-1",
                "aws_access_key_id" => "AWS_ACCESS_KEY_ID",
                "aws_secret_access_key" => "AWS_SECRET_ACCESS_KEY",
                "endpoint" => "endpoint_to_s3_compatible_blobstore",
                "use_iam_profile" => false,
                "backup" => {
                  "name" => "the_packages_bucket_backu",
                  "region" => "eu-west-2",
                }
              }
            }
          }

         expect { buckets_template.render(manifest) }.to(raise_error(RuntimeError, 'Invalid bucket configuration, my_packages_bucket is used as a source bucket and a backup bucket'))
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
          expect(config).to include("/var/vcap/jobs/bpm/bin/bpm run s3-unversioned-blobstore-backup-restorer")
        end
      end

      context 'when bpm is not enabled' do
        it 'does not template bpm' do
          config = restore_template.render("enabled" => true)
          expect(config).to include("restore")
          expect(config).not_to include("/var/vcap/jobs/bpm/bin/bpm run s3-unversioned-blobstore-backup-restorer")
        end
      end
    end
  end
end
