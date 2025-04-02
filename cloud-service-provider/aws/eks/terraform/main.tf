provider "aws" {
  region = var.region
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

data "aws_availability_zones" "available" {
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
data "aws_caller_identity" "current" {}

locals {
  vpc_cidr = "10.0.0.0/16"
  azs      = slice(data.aws_availability_zones.available.names, 0, 3) 
}

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "${var.cluster_name}-vpc"
  cidr = local.vpc_cidr
  azs  = local.azs
  
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.4.0/24", "10.0.5.0/24", "10.0.6.0/24"]

  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  public_subnet_tags = {
    "kubernetes.io/role/elb" = 1
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
  }
}

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "20.8.5"

  cluster_name = var.cluster_name
  cluster_version = var.cluster_version
  
  cluster_endpoint_public_access           = true
  enable_cluster_creator_admin_permissions = true
  
  cluster_addons = {
    coredns                = { most_recent = true }
    kube-proxy             = { most_recent = true }
    vpc-cni                = { most_recent = true }
    aws-efs-csi-driver     = { most_recent = true }
  }

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  eks_managed_node_group_defaults = {
    instance_types = var.instance_types
    capacity_type  = var.capacity_type
    min_size       = var.min_size
    max_size       = var.max_size
    desired_size   = var.desired_size
    block_device_mappings = {
      xvda = {
        device_name = "/dev/xvda"
        ebs         = {
          volume_size           = var.disk_size
          volume_type           = "gp3"
          iops                  = 3000
          throughput            = 125
          encrypted             = true
          delete_on_termination = true
        }
      }
    }      
  }

  eks_managed_node_groups = {
   default = {
      iam_role_additional_policies = {
        AmazonEFSCSIDriverPolicy = "arn:aws:iam::aws:policy/service-role/AmazonEFSCSIDriverPolicy"
      }
    }
  }
}

module "efs" {
  source = "terraform-aws-modules/efs/aws"

  # File system
  name           = var.cluster_name
  creation_token = var.cluster_name
  encrypted      = true
  kms_key_arn    = module.kms.key_arn

  # File system policy
  policy_statements = [
    {
      sid     = "Example"
      actions = ["elasticfilesystem:ClientMount"]
      principals = [
        {
          type        = "AWS"
          identifiers = [data.aws_caller_identity.current.arn]
        }
      ]
    }
  ]

  # Mount targets / security group
  mount_targets              = { for k, v in zipmap(local.azs, module.vpc.private_subnets) : k => { subnet_id = v } }
  security_group_description = "EFS security group"
  security_group_vpc_id      = module.vpc.vpc_id
  security_group_rules = {
    vpc = {
      description = "NFS ingress from VPC private subnets"
      cidr_blocks = module.vpc.private_subnets_cidr_blocks
    }
  }

  tags = {
    Terraform   = "true"
    Environment = "dev"
  }
}

module "kms" {
  source  = "terraform-aws-modules/kms/aws"

  aliases               = ["efs/${var.cluster_name}"]
  description           = "EFS customer managed key"
  enable_default_policy = true

  # For example use only
  deletion_window_in_days = 7
}

resource "null_resource" "kubectl" {
    provisioner "local-exec" {
        command = "aws eks --region ${var.region} update-kubeconfig --name ${var.cluster_name}"
    }
    depends_on = [ module.eks ]
}

resource "kubernetes_storage_class_v1" "eks_efs" {
  metadata {
    name = "efs-sc"
  }
  storage_provisioner = "efs.csi.aws.com"
  reclaim_policy      = "Retain"
  parameters = {
    provisioningMode = "efs-ap"
    fileSystemId = "${module.efs.id}"
    directoryPerms = "700"
  }
  depends_on = [ 
    null_resource.kubectl
   ]
}
