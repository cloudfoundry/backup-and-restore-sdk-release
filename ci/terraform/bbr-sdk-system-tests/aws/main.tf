terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

# Passwords are generated once by terraform and kept in the S3 state file.
# They are exposed as outputs so the pipeline can read them without a
# separate CredHub entry. No pipeline credential is required for these.
resource "random_password" "postgres_13" {
  length  = 20
  special = false
}

resource "random_password" "postgres_15" {
  length  = 20
  special = false
}

resource "random_password" "postgres_16" {
  length  = 20
  special = false
}

resource "random_password" "mariadb_10" {
  length  = 20
  special = false
}

variable "aws_access_key" {
  type = string
}

variable "aws_secret_key" {
  type = string
}

variable "aws_assumed_role_arn" {
  type    = string
  default = ""
}

variable "aws_region" {
  type = string
}

resource "aws_db_instance" "backup_and_restore_postgres_13" {
  identifier          = "postgres-13-system-tests"
  allocated_storage   = 20
  storage_type        = "gp2"
  engine              = "postgres"
  engine_version      = "13"
  instance_class      = "db.t3.micro"
  username            = "root"
  password            = random_password.postgres_13.result
  publicly_accessible = true
  skip_final_snapshot = true
}

resource "aws_db_instance" "backup_and_restore_postgres_15" {
  identifier          = "postgres-15-system-tests"
  allocated_storage   = 20
  storage_type        = "gp2"
  engine              = "postgres"
  engine_version      = "15"
  instance_class      = "db.t3.micro"
  username            = "root"
  password            = random_password.postgres_15.result
  publicly_accessible = true
  skip_final_snapshot = true
}

resource "aws_db_instance" "backup_and_restore_postgres_16" {
  identifier          = "postgres-16-system-tests"
  allocated_storage   = 20
  storage_type        = "gp2"
  engine              = "postgres"
  engine_version      = "16"
  instance_class      = "db.t3.micro"
  username            = "root"
  password            = random_password.postgres_16.result
  publicly_accessible = true
  skip_final_snapshot = true
}

resource "aws_db_instance" "backup_and_restore_mariadb_10_6" {
  identifier          = "mariadb-10-6-system-tests"
  allocated_storage   = 20
  storage_type        = "gp2"
  engine              = "mariadb"
  engine_version      = "10.6"
  instance_class      = "db.t3.micro"
  username            = "root"
  password            = random_password.mariadb_10.result
  publicly_accessible = true
  skip_final_snapshot = true
}

output "postgres_13_address" {
  value = aws_db_instance.backup_and_restore_postgres_13.address
}

output "postgres_15_address" {
  value = aws_db_instance.backup_and_restore_postgres_15.address
}

output "postgres_16_address" {
  value = aws_db_instance.backup_and_restore_postgres_16.address
}

output "mariadb_10_6_address" {
  value = aws_db_instance.backup_and_restore_mariadb_10_6.address
}

# Passwords are exposed as sensitive outputs so the pipeline can
# pass them to test tasks without storing them in CredHub.
output "postgres_13_password" {
  value     = random_password.postgres_13.result
  sensitive = true
}

output "postgres_15_password" {
  value     = random_password.postgres_15.result
  sensitive = true
}

output "postgres_16_password" {
  value     = random_password.postgres_16.result
  sensitive = true
}

output "mariadb_10_password" {
  value     = random_password.mariadb_10.result
  sensitive = true
}

provider "aws" {
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = var.aws_region

  dynamic "assume_role" {
    for_each = (length(var.aws_assumed_role_arn) > 0) ? toset([1]) : toset([])
    content {
      role_arn = var.aws_assumed_role_arn
    }
  }
}
