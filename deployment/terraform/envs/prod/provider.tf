terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.17.0"
    }

    sops = {
      source = "carlpett/sops"
      version = "1.0.0"
    }
  }

  backend "s3" {
    bucket               = "soerenschneider-terraform"
    key                  = "dyndns-server"
    region               = "us-east-1"
  }
}

provider "aws" {
  region = "us-east-1"
}

