require 'rspec'
require 'yaml'
require 'json'
require 'bosh/template/test'

describe 'gcs-blobstore-backup-restorer job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../..')) }
  let(:job) { release.job('gcs-blobstore-backup-restorer') }
  let(:backup_template) { job.template('bin/bbr/backup') }
  let(:metadata_template) { job.template('bin/bbr/metadata') }
  let(:buckets_template) { job.template('config/buckets.json') }
  let(:gcp_service_account_key_template) { job.template('config/gcp-service-account-key.json') }
  let(:restore_template) { job.template('bin/bbr/restore') }

  describe 'backuper templates' do
    context 'when backup is not enabled' do
      it 'the templated script is empty' do
        config = backup_template.render({})
        expect(config.strip).to eq("#!/usr/bin/env bash\n\nset -eu")
      end

      it 'the bucket config is empty' do
        manifest = {
          "enabled" => false,
          "gcp_service_account_key" => nil,
          "buckets" => {
            "droplets"  => nil
          }
        }
        config = buckets_template.render(manifest)
        expect(config.strip).to eq("")
      end

      it 'the gcp_service_account_key config is empty' do
        manifest = {
          "enabled" => false,
          "gcp_service_account_key" => nil,
          "buckets" => {
            "droplets"  => nil
          }
        }
        config = gcp_service_account_key_template.render(manifest)
        expect(config.strip).to eq("")
      end

      it 'the metadata script enables the skip_bbr_scripts flag' do
        metadata = metadata_template.render({})
        expect(metadata).to include("skip_bbr_scripts: true")
      end
    end

    context 'when backup is enabled' do
      it 'the metadata script disables the skip_bbr_scripts flag' do
        metadata = metadata_template.render({"enabled" => true})
        expect(metadata).to include("skip_bbr_scripts: false")
      end

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
            {
              "enabled" => true,
              "buckets" => {"droplets" => {
                  "bucket_name" => "my_bucket",
                  "backup_bucket_name" => "my_bucket",
                  "gcp_service_account_key" => "{}"
                  }
                }
            }
          ) }.to(raise_error(RuntimeError, "Invalid bucket configuration for 'droplets', bucket_name and backup_bucket_name must be distinct"))
        end
      end

      context 'and the bucket_id is blank' do
        it 'errors' do
          manifest = {
            "enabled" => true,
            "buckets" => {
              "    " => {
                "bucket_name" => "my_bucket",
                "backup_bucket_name" => "my_backup_bucket",
              }
            }
          }
          expect { buckets_template.render(manifest) }.to(raise_error(
            RuntimeError, "Invalid buckets configuration, must not be blank"
          ))
        end
      end

      context 'and the live bucket name is blank' do
        it 'errors' do
          manifest = {
            "enabled" => true,
            "buckets" => {
              "droplets" => {
                "bucket_name" => "     ",
                "backup_bucket_name" => "my_backup_bucket",
              }
            }
          }
          expect { buckets_template.render(manifest) }.to(raise_error(
            RuntimeError, "Invalid bucket configuration for 'droplets', bucket_name and backup_bucket_name must be configured"
          ))
        end
      end

      context 'and the backup bucket name is blank' do
        it 'errors' do
          manifest = {
            "enabled" => true,
            "buckets" => {
              "droplets" => {
                "bucket_name" => "my_bucket",
                "backup_bucket_name" => " ",
              }
            }
          }
          expect { buckets_template.render(manifest) }.to(raise_error(
            RuntimeError, "Invalid bucket configuration for 'droplets', bucket_name and backup_bucket_name must be configured"
          ))
        end
      end

      context 'and the buckets are empty hash' do
        it 'errors' do
          manifest = {
            "enabled" => true,
            "buckets" => {
              "droplets" => {}
            }
          }
          expect { buckets_template.render(manifest) }.to(raise_error(
            RuntimeError, "Invalid bucket configuration for 'droplets', bucket_name and backup_bucket_name must be configured"
          ))
        end
      end

      context 'the GCS key provided is not valid JSON' do
        it 'errors' do
          expect { gcp_service_account_key_template.render(
            {
              "enabled" => true,
              "gcp_service_account_key" => "{not valid json}"
            }
          ) }.to(raise_error(RuntimeError, 'Invalid gcp_service_account_key provided; it is not valid JSON'))
        end
      end

      context 'and the backup bucket is the same as a different live bucket' do
        it 'errors' do
          expect { buckets_template.render(
            {
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
            }
          ) }.to(raise_error(RuntimeError, "Invalid bucket configuration, 'bucket2' is used as a source bucket and a backup bucket"))
        end
      end
    end
  end
end
