terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.99.1"
    }

    sops = {
      source  = "carlpett/sops"
      version = "1.2.1"
    }
  }

  backend "s3" {
    bucket = "soerenschneider-terraform"
    key    = "dyndns-server"
    region = "us-east-1"
  }
}

provider "aws" {
  region = "us-east-1"
}

