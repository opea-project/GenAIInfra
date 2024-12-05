## 🤖 Terraform Example for AWS EC2 VM using the default VPC

This example uses the following Terraform module:

[Terraform Module](https://registry.terraform.io/modules/intel/aws-vm/intel/latest)

**For additional customization, refer to the module documentation under "Inputs".**

**The Module supports non-default VPCs and much more than what is shown in this example.**

## Overview

This example creates an AWS EC2 in the default VPC. The default region is can be changed in variables.tf.

This example also creates:

- Public IP
- EC2 key pair
- The private key is created in the local system where terraform apply is done
- It also creates a new security groups for network access

## Prerequisites

1. **Install AWS CLI**: Follow the instructions to install the AWS CLI from the [official documentation](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html).

2. **Configure AWS CLI**: Run the following command to configure your AWS credentials and default region.
3. **Install Terraform**: Follow the instructions to install Terraform from the [official documentation](https://learn.hashicorp.com/tutorials/terraform/install-cli).
4. Have your preferred Git client installed. If you don't have one, you can download it from [Git](https://git-scm.com/downloads).

## Configure AWS CLI

```bash
aws configure
```

You will be prompted to enter your AWS Access Key ID, Secret Access Key, default region name, and output format.

## Modify the example to suit your needs

This example is can be customized to your needs by modifying the following files:

```bash
main.tf
variables.tf
```

For additional customization, refer to the module documentation under **"Inputs"** [Terraform Module](https://registry.terraform.io/modules/intel/aws-vm/intel/latest)
.

The module supports much more than what is shown in this example.

## Usage

In variables.tf, replace the below with you own IPV4 CIDR range before running the example.

Use <https://whatismyipaddress.com/> to get your IP address.

```hcl
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = "A.B.C.D/32"
```

**Depending on your use case, you might also need to allow additional.
ports.**

**Modify variable.tf replicating the existing format to add additional ports.**

## Run Terraform

```bash
git clone https://github.com/opea-project/GenAIInfra.git
cd GenAIInfra/cloud-service-provider/aws/ec2-vm/terraform

# Modify the main.tf and variables.tf file to suit your needs (see above)

terraform init
terraform plan
terraform apply
```

## SSH

At this point, the EC2 instance should be up and running. You can SSH into the instance using the private key created in the local system.

```bash
chmod 600 tfkey.private
ssh -i tfkey.private ubuntu@***VM_PUBLIC_IP***

# If in a proxy environment, use the following command
ssh -i tfkey.private -x your-proxy.com.com:PORT ubuntu@***VM_PUBLIC_IP***
```

## OPEA

You can now deploy OPEA components using OPEA instructions.

[OPEA GenAI Examples](https://github.com/opea-project/GenAIExamples)

## Destroy

To destroy the resources created by this example, run the following command:

```bash
terraform destroy
```

## Considerations

The AWS region where this example is run should have a default VPC.
