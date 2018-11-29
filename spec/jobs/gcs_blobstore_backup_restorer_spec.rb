require 'rspec'
require 'yaml'
require 'json'
require 'bosh/template/test'

describe 'gcs-blobstore-backup-restorer job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../..')) }
  let(:job) { release.job('gcs-blobstore-backup-restorer') }
  let(:backup_template) { job.template('bin/bbr/backup') }
  let(:buckets_template) { job.template('config/buckets.json') }
  let(:gcp_service_account_key_template) { job.template('config/gcp-service-account-key.json') }
  let(:restore_template) { job.template('bin/bbr/restore') }

  describe 'backup' do
    context 'when backup is not enabled' do
      it 'the templated script is empty' do
        config = backup_template.render({})
        expect(config.strip).to eq("#!/usr/bin/env bash\n\nset -eu")
      end
    end

    context 'when backup is enabled' do
      context 'and it is configured correctly' do
        it 'succeeds' do
          manifest = {
            "enabled" => true,
            "buckets" => {"droplets" => {
                "bucket_name" => "my_bucket",
                "backup_bucket_name" => "my_backup_bucket",
                }
              },
            "gcp_service_account_key" => "{}"
          }
          expect { buckets_template.render(manifest) }.to_not(raise_error)
          expect { gcp_service_account_key_template.render(manifest) }.to_not(raise_error)
        end
      end

      context 'and the backup bucket is the same as the live bucket' do
        it 'errors' do
          expect { buckets_template.render(
            "enabled" => true,
            "buckets" => {"droplets" => {
                "bucket_name" => "my_bucket",
                "backup_bucket_name" => "my_bucket",
                "gcp_service_account_key" => "{}"
                }
              }
          ) }.to(raise_error(RuntimeError, 'Invalid bucket configuration for droplets, bucket_name and backup_bucket_name must be distinct'))
        end
      end

      context 'the GCS key provided is not valid JSON' do
        it 'errors' do
          expect { gcp_service_account_key_template.render(
            "gcp_service_account_key" => "{not valid json}"
          ) }.to(raise_error(RuntimeError, 'Invalid gcp_service_account_key provided; it is not valid JSON'))
        end
      end

      context 'and the backup bucket is the same as a different live bucket' do
        it 'errors' do
          expect { buckets_template.render(
            "enabled" => true,
            "buckets" => {"droplets" => {
                "bucket_name" => "bucket1",
                "backup_bucket_name" => "bucket2",
                "gcp_service_account_key" => "{}"
              },
              "packages" => {
                "bucket_name" => "bucket2",
                "backup_bucket_name" => "bucket3",
                "gcp_service_account_key" => "{}"
              }
            }
          ) }.to(raise_error(RuntimeError, 'Invalid bucket configuration, bucket2 is used as a source bucket and a backup bucket'))
        end
      end
    end
  end
end
