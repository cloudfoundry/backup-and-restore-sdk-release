#!/usr/bin/env bash
#
# Creates the IAM resources the backup-and-restore-sdk-release pipeline needs
# in the AWS account 504405067906.
#
#   bosh_bbr_sdk_pipeline            IAM user. Its only permission is
#                                    assuming the role below. Its access key
#                                    goes into CredHub as bbr_sdk_aws_creds.
#   bosh_bbr_sdk_pipeline_role       Role the pipeline and tests assume. The
#                                    tests require a role ARN
#                                    (S3_ASSUMED_ROLE_ARN, AWS_ASSUMED_ROLE_ARN).
#   ci-bbr-sdk-bbl-director          Role + instance profile for the director
#                                    bbl creates for the IAM instance-profile
#                                    test. bbl uses it instead of creating its
#                                    own, so the pipeline role needs no IAM
#                                    write permissions.
#   ci-bbr-sdk-s3-backuper           Role + instance profile attached to the
#                                    s3-backuper VM in the IAM instance-profile
#                                    test. S3 access to the IAM test buckets only.
#
# Safe to re-run: existing users/roles/profiles are kept, and policies are
# (re)applied. The access key is only created when the user has none.
#
# Usage:
#   ./ci/scripts/create-aws-iam.sh
#   DRY_RUN=true ./ci/scripts/create-aws-iam.sh   # print commands only

set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

ACCOUNT_ID="504405067906"
REGIONS='["eu-north-1", "eu-west-1"]'
DRY_RUN="${DRY_RUN:-false}"

USER_NAME="bosh_bbr_sdk_pipeline"
PIPELINE_ROLE="bosh_bbr_sdk_pipeline_role"
DIRECTOR_ROLE="ci-bbr-sdk-bbl-director"
DIRECTOR_PROFILE="ci-bbr-sdk-bbl-director"
BACKUPER_ROLE="ci-bbr-sdk-s3-backuper"
BACKUPER_PROFILE="ci-bbr-sdk-s3-backuper"

MAX_SESSION_SECONDS=14400 # 4h: longer than any task that assumes the role

run() {
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "+ $*"
  else
    "$@"
  fi
}

check_account() {
  local account_id
  account_id="$(aws sts get-caller-identity --query Account --output text)"
  if [[ "${account_id}" != "${ACCOUNT_ID}" ]]; then
    echo "ERROR: logged in to AWS account ${account_id}, expected ${ACCOUNT_ID}" >&2
    exit 1
  fi
}

user_policy() {
  cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AssumePipelineRole",
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Resource": "arn:aws:iam::${ACCOUNT_ID}:role/${PIPELINE_ROLE}"
    }
  ]
}
EOF
}

pipeline_role_trust_policy() {
  cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "AWS": "arn:aws:iam::${ACCOUNT_ID}:user/${USER_NAME}" },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

pipeline_role_policy() {
  cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "S3PipelineAndTestBuckets",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::ci-bbr-sdk-*",
        "arn:aws:s3:::ci-bbr-sdk-*/*",
        "arn:aws:s3:::sdk-integration-test-*",
        "arn:aws:s3:::sdk-integration-test-*/*"
      ]
    },
    {
      "Sid": "S3ListBuckets",
      "Effect": "Allow",
      "Action": "s3:ListAllMyBuckets",
      "Resource": "*"
    },
    {
      "Sid": "RdsSystemTestInstances",
      "Effect": "Allow",
      "Action": "rds:*",
      "Resource": [
        "arn:aws:rds:*:${ACCOUNT_ID}:db:*-system-tests",
        "arn:aws:rds:*:${ACCOUNT_ID}:pg:*",
        "arn:aws:rds:*:${ACCOUNT_ID}:og:*",
        "arn:aws:rds:*:${ACCOUNT_ID}:subgrp:*"
      ],
      "Condition": { "StringEquals": { "aws:RequestedRegion": ${REGIONS} } }
    },
    {
      "Sid": "RdsDescribe",
      "Effect": "Allow",
      "Action": ["rds:Describe*", "rds:ListTagsForResource"],
      "Resource": "*"
    },
    {
      "Sid": "BblInfrastructure",
      "Effect": "Allow",
      "Action": ["ec2:*", "elasticloadbalancing:*", "logs:*", "kms:*"],
      "Resource": "*",
      "Condition": { "StringEquals": { "aws:RequestedRegion": ${REGIONS} } }
    },
    {
      "Sid": "LaunchBblDirector",
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "arn:aws:iam::${ACCOUNT_ID}:role/${DIRECTOR_ROLE}"
    },
    {
      "Sid": "IamReadOnly",
      "Effect": "Allow",
      "Action": ["iam:List*", "iam:Get*"],
      "Resource": "*"
    }
  ]
}
EOF
}

# bbl's default director policy, except that the director can only pass the
# backuper role, and can't create service-linked roles.
director_policy() {
  cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Cpi",
      "Effect": "Allow",
      "Action": [
        "ec2:AssociateAddress",
        "ec2:AttachVolume",
        "ec2:CancelSpotInstanceRequests",
        "ec2:CopyImage",
        "ec2:CreateNetworkInterface",
        "ec2:CreateRoute",
        "ec2:CreateSnapshot",
        "ec2:CreateTags",
        "ec2:CreateVolume",
        "ec2:DeleteNetworkInterface",
        "ec2:DeleteSnapshot",
        "ec2:DeleteVolume",
        "ec2:DeregisterImage",
        "ec2:DescribeAddresses",
        "ec2:DescribeAvailabilityZones",
        "ec2:DescribeImages",
        "ec2:DescribeInstanceTypes",
        "ec2:DescribeInstances",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DescribeRegions",
        "ec2:DescribeRouteTables",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSnapshots",
        "ec2:DescribeSpotInstanceRequests",
        "ec2:DescribeSubnets",
        "ec2:DescribeVolumes",
        "ec2:DetachVolume",
        "ec2:ModifyInstanceAttribute",
        "ec2:ModifyNetworkInterfaceAttribute",
        "ec2:RegisterImage",
        "ec2:ReplaceRoute",
        "ec2:RequestSpotInstances",
        "ec2:RunInstances",
        "ec2:TerminateInstances",
        "elasticloadbalancing:*",
        "kms:CreateGrant",
        "kms:DescribeKey*",
        "kms:GenerateDataKey*",
        "kms:ReEncrypt*"
      ],
      "Resource": "*",
      "Condition": { "StringEquals": { "aws:RequestedRegion": ${REGIONS} } }
    },
    {
      "Sid": "PassBackuperRole",
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "arn:aws:iam::${ACCOUNT_ID}:role/${BACKUPER_ROLE}"
    }
  ]
}
EOF
}

ec2_trust_policy() {
  cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Service": "ec2.amazonaws.com" },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

backuper_policy() {
  cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "IamTestBuckets",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::ci-bbr-sdk-iam-instance-role-test",
        "arn:aws:s3:::ci-bbr-sdk-iam-instance-role-test/*",
        "arn:aws:s3:::ci-bbr-sdk-iam-instance-role-test-clone",
        "arn:aws:s3:::ci-bbr-sdk-iam-instance-role-test-clone/*"
      ]
    }
  ]
}
EOF
}

ensure_user() {
  if aws iam get-user --user-name "${USER_NAME}" >/dev/null 2>&1; then
    echo "exists:  user ${USER_NAME}"
  else
    echo "create:  user ${USER_NAME}"
    run aws iam create-user --user-name "${USER_NAME}" >/dev/null
  fi
  run aws iam put-user-policy \
    --user-name "${USER_NAME}" \
    --policy-name assume-pipeline-role \
    --policy-document "$(user_policy)"
}

ensure_role() {
  local role="$1" trust="$2" policy_name="$3" policy="$4"

  if aws iam get-role --role-name "${role}" >/dev/null 2>&1; then
    echo "exists:  role ${role}"
    run aws iam update-assume-role-policy --role-name "${role}" --policy-document "${trust}"
  else
    echo "create:  role ${role}"
    # A just-created user can take a few seconds to be accepted as a trust
    # policy principal ("Invalid principal in policy"), so retry.
    local attempt
    for attempt in 1 2 3 4 5 6; do
      if run aws iam create-role \
        --role-name "${role}" \
        --assume-role-policy-document "${trust}" \
        --max-session-duration "${MAX_SESSION_SECONDS}" \
        --query Role.Arn --output text; then
        break
      fi
      if [[ "${attempt}" == 6 ]]; then
        echo "ERROR: could not create role ${role}" >&2
        exit 1
      fi
      sleep 10
    done
  fi
  run aws iam put-role-policy \
    --role-name "${role}" \
    --policy-name "${policy_name}" \
    --policy-document "${policy}"
}

ensure_instance_profile() {
  local profile="$1" role="$2"

  if aws iam get-instance-profile --instance-profile-name "${profile}" >/dev/null 2>&1; then
    echo "exists:  instance profile ${profile}"
  else
    echo "create:  instance profile ${profile}"
    run aws iam create-instance-profile --instance-profile-name "${profile}" >/dev/null
  fi

  local attached
  attached="$(aws iam get-instance-profile --instance-profile-name "${profile}" \
    --query 'InstanceProfile.Roles[].RoleName' --output text 2>/dev/null || true)"
  if [[ "${attached}" != *"${role}"* ]]; then
    run aws iam add-role-to-instance-profile \
      --instance-profile-name "${profile}" \
      --role-name "${role}"
  fi
}

ensure_access_key() {
  local keys
  keys="$(aws iam list-access-keys --user-name "${USER_NAME}" \
    --query 'AccessKeyMetadata[].AccessKeyId' --output text 2>/dev/null || true)"
  if [[ -n "${keys}" && "${keys}" != "None" ]]; then
    echo "exists:  access key(s) for ${USER_NAME}: ${keys} (not creating another)"
    return
  fi

  echo "create:  access key for ${USER_NAME}"
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "+ aws iam create-access-key --user-name ${USER_NAME}"
    return
  fi

  local key
  key="$(aws iam create-access-key --user-name "${USER_NAME}" \
    --query 'AccessKey.[AccessKeyId,SecretAccessKey]' --output text)"
  local key_id secret
  read -r key_id secret <<< "${key}"

  cat <<EOF

Store this in CredHub (the secret is shown only once):

  credhub set -t json -n /concourse/main/backup-and-restore-sdk-release/bbr_sdk_aws_creds -v '{
    "access_key_id": "${key_id}",
    "secret_access_key": "${secret}",
    "role_arn": "arn:aws:iam::${ACCOUNT_ID}:role/${PIPELINE_ROLE}"
  }'

The pipeline's iam-instance-profile anchor must be ${BACKUPER_PROFILE}, and
its bbl-director-instance-profile anchor ${DIRECTOR_PROFILE}.
EOF
}

main() {
  check_account

  ensure_user
  ensure_role "${PIPELINE_ROLE}" "$(pipeline_role_trust_policy)" pipeline "$(pipeline_role_policy)"
  ensure_role "${BACKUPER_ROLE}" "$(ec2_trust_policy)" iam-test-buckets "$(backuper_policy)"
  ensure_instance_profile "${BACKUPER_PROFILE}" "${BACKUPER_ROLE}"
  ensure_role "${DIRECTOR_ROLE}" "$(ec2_trust_policy)" bbl-director "$(director_policy)"
  ensure_instance_profile "${DIRECTOR_PROFILE}" "${DIRECTOR_ROLE}"
  ensure_access_key
}

main "$@"
