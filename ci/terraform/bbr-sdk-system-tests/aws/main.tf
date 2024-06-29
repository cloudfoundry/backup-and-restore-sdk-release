variable "postgres_11_password" {
    type = string
}

variable "postgres_13_password" {
    type = string
}

variable "postgres_15_password" {
    type = string
}

variable "mariadb_10_password" {
    type = string
}

variable "aws_access_key" {
    type = string
}

variable "aws_secret_key" {
    type = string
}

variable "aws_assumed_role_arn" {
    type = string
    default = ""
}

variable "aws_region" {
    type = string
}

resource "aws_db_instance" "backup_and_restore_postgres_11" {
  identifier           = "postgres-11-system-tests"
  allocated_storage    = 20
  storage_type         = "gp2"
  engine               = "postgres"
  engine_version       = "11"
  instance_class       = "db.t3.micro"
  username             = "root"
  password             = var.postgres_11_password
  publicly_accessible  = true
  skip_final_snapshot  = true
}

resource "aws_db_instance" "backup_and_restore_postgres_13" {
  identifier           = "postgres-13-system-tests"
  allocated_storage    = 20
  storage_type         = "gp2"
  engine               = "postgres"
  engine_version       = "13"
  instance_class       = "db.t3.micro"
  username             = "root"
  password             = var.postgres_13_password
  publicly_accessible  = true
  skip_final_snapshot  = true
}

resource "aws_db_instance" "backup_and_restore_postgres_15" {
  identifier           = "postgres-15-system-tests"
  allocated_storage    = 20
  storage_type         = "gp2"
  engine               = "postgres"
  engine_version       = "15"
  instance_class       = "db.t3.micro"
  username             = "root"
  password             = var.postgres_15_password
  publicly_accessible  = true
  skip_final_snapshot  = true
}

resource "aws_db_instance" "backup_and_restore_mariadb_10_6" {
  identifier           = "mariadb-10-6-system-tests"
  allocated_storage    = 20
  storage_type         = "gp2"
  engine               = "mariadb"
  engine_version       = "10.6"
  instance_class       = "db.t3.micro"
  username             = "root"
  password             = var.mariadb_10_password
  publicly_accessible  = true
  skip_final_snapshot  = true
}

output "postgres_11_address" {
  value = aws_db_instance.backup_and_restore_postgres_11.address
}
output "postgres_13_address" {
  value = aws_db_instance.backup_and_restore_postgres_13.address
}
output "postgres_15_address" {
  value = aws_db_instance.backup_and_restore_postgres_15.address
}

output "mariadb_10_6_address" {
  value = aws_db_instance.backup_and_restore_mariadb_10_6.address
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
