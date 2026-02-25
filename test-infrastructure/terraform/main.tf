terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# Variables
variable "aws_region" {
  description = "AWS region"
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  default     = "hbf-test"
}

variable "key_name" {
  description = "SSH key pair name"
  type        = string
}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.environment}-vpc"
    Environment = var.environment
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name        = "${var.environment}-igw"
    Environment = var.environment
  }
}

# Subnets
resource "aws_subnet" "management" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.0.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name        = "${var.environment}-management-subnet"
    Environment = var.environment
  }
}

resource "aws_subnet" "application" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.1.0.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name        = "${var.environment}-application-subnet"
    Environment = var.environment
  }
}

resource "aws_subnet" "database" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.2.0.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = false

  tags = {
    Name        = "${var.environment}-database-subnet"
    Environment = var.environment
  }
}

# Route Tables
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name        = "${var.environment}-public-rt"
    Environment = var.environment
  }
}

resource "aws_route_table_association" "management" {
  subnet_id      = aws_subnet.management.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "application" {
  subnet_id      = aws_subnet.application.id
  route_table_id = aws_route_table.public.id
}

# Security Groups
resource "aws_security_group" "management" {
  name        = "${var.environment}-management-sg"
  description = "Security group for management services"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8500
    to_port     = 8500
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 9090
    to_port     = 9090
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.environment}-management-sg"
    Environment = var.environment
  }
}

resource "aws_security_group" "hbf_nodes" {
  name        = "${var.environment}-hbf-nodes-sg"
  description = "Security group for HBF agent nodes"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8080
    to_port     = 8081
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 9090
    to_port     = 9092
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.environment}-hbf-nodes-sg"
    Environment = var.environment
  }
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
}

# EC2 Instances

# Consul Server
resource "aws_instance" "consul" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.medium"
  subnet_id              = aws_subnet.management.id
  vpc_security_group_ids = [aws_security_group.management.id]
  key_name               = var.key_name
  private_ip             = "10.0.0.10"

  user_data = file("${path.module}/user-data/consul.sh")

  tags = {
    Name        = "${var.environment}-consul"
    Environment = var.environment
    Role        = "consul"
  }
}

# Prometheus Server
resource "aws_instance" "prometheus" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.medium"
  subnet_id              = aws_subnet.management.id
  vpc_security_group_ids = [aws_security_group.management.id]
  key_name               = var.key_name
  private_ip             = "10.0.0.11"

  user_data = file("${path.module}/user-data/prometheus.sh")

  tags = {
    Name        = "${var.environment}-prometheus"
    Environment = var.environment
    Role        = "prometheus"
  }
}

# Grafana Server
resource "aws_instance" "grafana" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.small"
  subnet_id              = aws_subnet.management.id
  vpc_security_group_ids = [aws_security_group.management.id]
  key_name               = var.key_name
  private_ip             = "10.0.0.12"

  user_data = file("${path.module}/user-data/grafana.sh")

  tags = {
    Name        = "${var.environment}-grafana"
    Environment = var.environment
    Role        = "grafana"
  }
}

# HBF Application Nodes
resource "aws_instance" "hbf_nodes" {
  count                  = 3
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.medium"
  subnet_id              = aws_subnet.application.id
  vpc_security_group_ids = [aws_security_group.hbf_nodes.id]
  key_name               = var.key_name
  private_ip             = "10.1.0.${101 + count.index}"

  user_data = templatefile("${path.module}/user-data/hbf-node.sh", {
    node_id     = "hbf-node-${count.index + 1}"
    consul_addr = aws_instance.consul.private_ip
  })

  tags = {
    Name        = "${var.environment}-hbf-node-${count.index + 1}"
    Environment = var.environment
    Role        = "hbf-agent"
  }
}

# HBF Database Nodes
resource "aws_instance" "db_nodes" {
  count                  = 3
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.large"
  subnet_id              = aws_subnet.database.id
  vpc_security_group_ids = [aws_security_group.hbf_nodes.id]
  key_name               = var.key_name
  private_ip             = "10.2.0.${201 + count.index}"

  user_data = templatefile("${path.module}/user-data/db-node.sh", {
    node_id     = "db-node-${count.index + 1}"
    consul_addr = aws_instance.consul.private_ip
  })

  tags = {
    Name        = "${var.environment}-db-node-${count.index + 1}"
    Environment = var.environment
    Role        = "database"
  }
}

# Outputs
output "consul_public_ip" {
  value = aws_instance.consul.public_ip
}

output "prometheus_public_ip" {
  value = aws_instance.prometheus.public_ip
}

output "grafana_public_ip" {
  value = aws_instance.grafana.public_ip
}

output "hbf_nodes_public_ips" {
  value = aws_instance.hbf_nodes[*].public_ip
}

output "db_nodes_private_ips" {
  value = aws_instance.db_nodes[*].private_ip
}

output "consul_ui_url" {
  value = "http://${aws_instance.consul.public_ip}:8500"
}

output "prometheus_url" {
  value = "http://${aws_instance.prometheus.public_ip}:9090"
}

output "grafana_url" {
  value = "http://${aws_instance.grafana.public_ip}:3000"
}
