# Provision EC2 Instance on AWS in default vpc. It is configured to create the EC2 in
# US-East-1 region. The region is provided in variables.tf in this example folder.

# This example also create an EC2 key pair. Associate the public key with the EC2 instance. 
# Creates the private key in the local system where terraform apply is done. 
# Create a new security group to open up the SSH port 22 to a specific IP CIDR block
# To ssh: 
# chmod 600 tfkey.private
# ssh -i tfkey.private ubuntu@<public_ip>

data "aws_ami" "ubuntu-linux-2204" {
  most_recent = true
  owners      = ["099720109477"] # Canonical
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "random_id" "rid" {
  byte_length = 5
}

# RSA key of size 4096 bits
resource "tls_private_key" "rsa" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "TF_key" {
  key_name   = "TF_key-${random_id.rid.dec}"
  public_key = tls_private_key.rsa.public_key_openssh
}

resource "local_file" "TF_private_key" {
  content  = tls_private_key.rsa.private_key_pem
  filename = "tfkey.private"
}
resource "aws_security_group" "ssh_security_group" {
  description = "security group to configure ports for ssh"
  name_prefix = "ssh_security_group"
}

# Modify the `ingress_rules` variable in the variables.tf file to allow the required ports for your CIDR ranges
resource "aws_security_group_rule" "ingress_rules" {
  count             = length(var.ingress_rules)
  type              = "ingress"
  security_group_id = aws_security_group.ssh_security_group.id
  from_port         = var.ingress_rules[count.index].from_port
  to_port           = var.ingress_rules[count.index].to_port
  protocol          = var.ingress_rules[count.index].protocol
  cidr_blocks       = [var.ingress_rules[count.index].cidr_blocks]
}

resource "aws_network_interface_sg_attachment" "sg_attachment" {
  count                = length(module.ec2-vm)
  security_group_id    = aws_security_group.ssh_security_group.id
  network_interface_id = module.ec2-vm[count.index].primary_network_interface_id
}

# Modify the `vm_count` variable in the variables.tf file to create the required number of EC2 instances
module "ec2-vm" {
  count             = var.vm_count
  source            = "intel/aws-vm/intel"
  version           = "1.3.3"
  key_name          = aws_key_pair.TF_key.key_name
  instance_type     = var.instance_type # Modify the instance type as required for your AI needs
  availability_zone = var.availability_zone
  ami               = data.aws_ami.ubuntu-linux-2204.id

  # Size of VM disk in GB
  root_block_device = [{
    volume_size = var.volume_size
  }]

  tags = {
    Name = "opea-vm-${random_id.rid.dec}"
  }
}