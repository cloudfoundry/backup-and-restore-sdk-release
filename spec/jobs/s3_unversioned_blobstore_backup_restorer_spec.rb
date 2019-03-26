require 'rspec'
require 'yaml'
require 'json'
require 'bosh/template/test'
require 'open3'
require 'pp'

describe 's3-unversioned-blobstore-backup-restorer job' do
  let(:release_dir) { File.join(File.dirname(__FILE__), '../..') }
  let(:release) { Bosh::Template::Test::ReleaseDir.new(release_dir) }
  let(:job) { release.job('s3-unversioned-blobstore-backup-restorer') }
  let(:buckets_template) { job.template('config/buckets.json') }


  describe 'backup' do
    let(:backup_template) { job.template('bin/bbr/backup') }
    let(:post_backup_unlock_template) { job.template('bin/bbr/post-backup-unlock') }

    context 'when backup is not enabled' do
      it 'the templated backup script is empty' do
        backup_script = backup_template.render({})
        expect(backup_script.strip).to eq("#!/usr/bin/env bash\n\nset -eu")

        post_backup_unlock_script = post_backup_unlock_template.render({})
        expect(post_backup_unlock_script.strip).to eq("#!/usr/bin/env bash\n\nset -eu")
      end

      it 'the templated buckets script is empty' do
        manifest = {
          "enabled" => false,
          "buckets" => {
            "droplets"  => {
              "name" => nil,
              "region" => nil,
              "aws_access_key_id" => nil,
              "aws_secret_access_key" => nil,
              "endpoint" => nil,
              "use_iam_profile" => nil,
              "backup" => nil
            }
          }
        }

        bucket_file = buckets_template.render(manifest)
        expect(bucket_file.strip).to eq("")
      end
    end

    context 'when backup is enabled' do
      context 'and bpm is enabled' do
        it 'templates bpm command correctly' do
          backup_script = backup_template.render({"bpm" => {"enabled" => true}, "enabled" => true})
          expect(backup_script).to include("/var/vcap/jobs/bpm/bin/bpm run s3-unversioned-blobstore-backup-restorer")
          expect(backup_script).to include("-p backup")

          post_backup_unlock_script = post_backup_unlock_template.render({"bpm" => {"enabled" => true}, "enabled" => true})
          expect(post_backup_unlock_script).to include("/var/vcap/jobs/bpm/bin/bpm run s3-unversioned-blobstore-backup-restorer")
          expect(post_backup_unlock_script).to include("-p post-backup-unlock")
        end
      end

      context 'and bpm is not enabled' do
        it 'does not template bpm' do
          backup_script = backup_template.render("enabled" => true)
          expect(backup_script).to include("backup")
          expect(backup_script).not_to include("/var/vcap/jobs/bpm/bin/bpm run s3-unversioned-blobstore-backup-restorer")

          post_backup_unlock_script = post_backup_unlock_template.render("enabled" => true)
          expect(post_backup_unlock_script).to include("unversioned-backup-complete")
          expect(post_backup_unlock_script).not_to include("/var/vcap/jobs/bpm/bin/bpm run s3-unversioned-blobstore-backup-restorer")
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
    let(:restore_template) { job.template('bin/bbr/restore') }

    context 'when restore is not enabled' do
      it 'the templated script is empty' do
        restore_script = restore_template.render({})
        expect(restore_script.strip).to eq("#!/usr/bin/env bash\n\nset -eu")
      end
    end

    context 'when restore is enabled' do
      context 'and when bpm is enabled' do
        it 'templates bpm command correctly' do
          restore_script = restore_template.render({"bpm" => {"enabled" => true}, "enabled" => true})
          expect(restore_script).to include("/var/vcap/jobs/bpm/bin/bpm run s3-unversioned-blobstore-backup-restorer")
        end
      end

      context 'when bpm is not enabled' do
        it 'does not template bpm' do
          restore_script = restore_template.render("enabled" => true)
          expect(restore_script).to include("restore")
          expect(restore_script).not_to include("/var/vcap/jobs/bpm/bin/bpm run s3-unversioned-blobstore-backup-restorer")
        end
      end
    end
  end

  describe 'metadata' do
    let(:metadata_template) { job.template('bin/bbr/metadata') }
    let(:metadata_script) { metadata_template.render({}) }
    let(:metadata_cmd) { "bash -c '#{metadata_script}'" }

    it 'it enables the skip_bbr_scripts flag when the job is disabled' do
      expect(metadata_script).to include("skip_bbr_scripts: true")
    end

    it 'it disables the skip_bbr_scripts flag when the job is enabled' do
      metadata = metadata_template.render("enabled" => true)
      expect(metadata).to include("skip_bbr_scripts: false")
    end

    it 'fails when called with an old version of bbr that does not export BBR_VERSION' do
      _, error_str, status = Open3.capture3(metadata_cmd)
      expect(status.success?).not_to eq(true)
      expect(error_str).to include('Error: BBR_VERSION is not set, please ensure you are using the latest version of bbr')
    end

    it 'fails to run when bbr major version set to string' do
        stdout_str, _, status = Open3.capture3({'BBR_VERSION' => 'foo.bar.baz'}, metadata_cmd)
        expect(status.success?).not_to eq(true)
        expect(stdout_str).to include('Error: BBR version must be a valid semVer')
    end

    it 'fails to run when bbr minor version set to string' do
        stdout_str, _, status = Open3.capture3({'BBR_VERSION' => '1.bar.baz'}, metadata_cmd)
        expect(status.success?).not_to eq(true)
        expect(stdout_str).to include('Error: BBR version must be a valid semVer')
    end

    it 'returns error if the bbr_version is less than 1.4.0' do
      stdout_str, _, status = Open3.capture3({'BBR_VERSION' => '1.3.2'}, metadata_cmd)
      expect(status.success?).not_to eq(true)
      expect(stdout_str).to include('Error: BBR version must be 1.4.0 or greater')
    end

    it 'succesfully runs if the bbr_version is 1.4.0' do
        stdout_str, _, status = Open3.capture3({'BBR_VERSION' => '1.4.0'}, metadata_cmd)
        expect(status.success?).to eq(true)
        expect(stdout_str).to include('---')
    end

    it 'succesfully runs if the bbr_version is greater than 1.4.0' do
        stdout_str, _, status = Open3.capture3({'BBR_VERSION' => '2.3.4'}, metadata_cmd)
        expect(status.success?).to eq(true)
        expect(stdout_str).to include('---')
    end
  end
end


