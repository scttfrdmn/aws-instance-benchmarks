# AWS Instance Benchmarks - Terraform Deployment
# Provides reproducible infrastructure for running research computing benchmarks

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
  description = "AWS region for benchmarking"
  type        = string
  default     = "us-east-1"
}

variable "benchmark_suites" {
  description = "List of benchmark suites to enable"
  type        = list(string)
  default     = ["genomics", "ml", "climate", "hep", "chemistry", "cfd", "astronomy", "social_sciences", "digital_humanities", "environmental", "engineering", "medical"]
}

variable "instance_families" {
  description = "EC2 instance families to benchmark"
  type        = list(string)
  default     = ["m7i", "c7i", "r7i", "m7a", "c7a", "r7a", "m7g", "c7g", "r7g", "p4d", "g5", "hpc7g"]
}

variable "benchmark_vpc_cidr" {
  description = "CIDR block for benchmark VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "enable_spot_instances" {
  description = "Enable spot instances for cost optimization"
  type        = bool
  default     = true
}

variable "max_benchmark_duration" {
  description = "Maximum duration for benchmark runs (in minutes)"
  type        = number
  default     = 120
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "benchmark_ami" {
  most_recent = true
  owners      = ["amazon"]
  
  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

# S3 Bucket for storing benchmark results
resource "aws_s3_bucket" "benchmark_results" {
  bucket_prefix = "aws-instance-benchmarks-results-"
  
  tags = {
    Name        = "AWS Instance Benchmarks Results"
    Environment = "production"
    Project     = "aws-instance-benchmarks"
  }
}

resource "aws_s3_bucket_versioning" "benchmark_results" {
  bucket = aws_s3_bucket.benchmark_results.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "benchmark_results" {
  bucket = aws_s3_bucket.benchmark_results.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# VPC for isolated benchmark environment
resource "aws_vpc" "benchmark_vpc" {
  cidr_block           = var.benchmark_vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "aws-instance-benchmarks-vpc"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "benchmark_igw" {
  vpc_id = aws_vpc.benchmark_vpc.id

  tags = {
    Name = "aws-instance-benchmarks-igw"
  }
}

# Public subnets for benchmark instances
resource "aws_subnet" "benchmark_public" {
  count             = min(length(data.aws_availability_zones.available.names), 3)
  vpc_id            = aws_vpc.benchmark_vpc.id
  cidr_block        = cidrsubnet(var.benchmark_vpc_cidr, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  
  map_public_ip_on_launch = true

  tags = {
    Name = "aws-instance-benchmarks-public-${count.index + 1}"
  }
}

# Route table for public subnets
resource "aws_route_table" "benchmark_public_rt" {
  vpc_id = aws_vpc.benchmark_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.benchmark_igw.id
  }

  tags = {
    Name = "aws-instance-benchmarks-public-rt"
  }
}

resource "aws_route_table_association" "benchmark_public_rta" {
  count          = length(aws_subnet.benchmark_public)
  subnet_id      = aws_subnet.benchmark_public[count.index].id
  route_table_id = aws_route_table.benchmark_public_rt.id
}

# Security group for benchmark instances
resource "aws_security_group" "benchmark_sg" {
  name        = "aws-instance-benchmarks-sg"
  description = "Security group for benchmark instances"
  vpc_id      = aws_vpc.benchmark_vpc.id

  # SSH access
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "SSH access for benchmark management"
  }

  # HTTP access for monitoring
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = [var.benchmark_vpc_cidr]
    description = "HTTP access for monitoring"
  }

  # HTTPS access for monitoring
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [var.benchmark_vpc_cidr]
    description = "HTTPS access for monitoring"
  }

  # Inter-instance communication for MPI workloads
  ingress {
    from_port = 0
    to_port   = 65535
    protocol  = "tcp"
    self      = true
    description = "Inter-instance communication"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outbound traffic"
  }

  tags = {
    Name = "aws-instance-benchmarks-sg"
  }
}

# IAM role for benchmark instances
resource "aws_iam_role" "benchmark_instance_role" {
  name = "aws-instance-benchmarks-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "benchmark_instance_policy" {
  name = "aws-instance-benchmarks-policy"
  role = aws_iam_role.benchmark_instance_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.benchmark_results.arn,
          "${aws_s3_bucket.benchmark_results.arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "cloudwatch:PutMetricData",
          "cloudwatch:PutDashboard",
          "cloudwatch:GetMetricStatistics"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ec2:DescribeInstances",
          "ec2:DescribeInstanceTypes",
          "ec2:DescribeImages",
          "ec2:DescribeSpotPriceHistory"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "pricing:GetProducts",
          "pricing:DescribeServices"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_instance_profile" "benchmark_instance_profile" {
  name = "aws-instance-benchmarks-profile"
  role = aws_iam_role.benchmark_instance_role.name
}

# Launch template for benchmark instances
resource "aws_launch_template" "benchmark_template" {
  name_prefix   = "aws-instance-benchmarks-"
  image_id      = data.aws_ami.benchmark_ami.id
  instance_type = "m7i.large"  # Default, will be overridden

  vpc_security_group_ids = [aws_security_group.benchmark_sg.id]

  iam_instance_profile {
    name = aws_iam_instance_profile.benchmark_instance_profile.name
  }

  user_data = base64encode(templatefile("${path.module}/user-data.sh", {
    s3_bucket = aws_s3_bucket.benchmark_results.id
    region    = var.aws_region
  }))

  tag_specifications {
    resource_type = "instance"
    tags = {
      Name    = "aws-instance-benchmark"
      Project = "aws-instance-benchmarks"
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Auto Scaling Groups for different benchmark suites
resource "aws_autoscaling_group" "benchmark_asg" {
  for_each = toset(var.benchmark_suites)

  name                = "aws-instance-benchmarks-${each.key}"
  vpc_zone_identifier = aws_subnet.benchmark_public[*].id
  target_group_arns   = []
  health_check_type   = "EC2"
  health_check_grace_period = 300

  min_size         = 0
  max_size         = 20
  desired_capacity = 0

  launch_template {
    id      = aws_launch_template.benchmark_template.id
    version = "$Latest"
  }

  tag {
    key                 = "Name"
    value               = "aws-instance-benchmarks-${each.key}"
    propagate_at_launch = true
  }

  tag {
    key                 = "BenchmarkSuite"
    value               = each.key
    propagate_at_launch = true
  }

  tag {
    key                 = "Project"
    value               = "aws-instance-benchmarks"
    propagate_at_launch = true
  }
}

# CloudWatch Log Groups for benchmark logs
resource "aws_cloudwatch_log_group" "benchmark_logs" {
  for_each = toset(var.benchmark_suites)

  name              = "/aws/benchmark/${each.key}"
  retention_in_days = 30

  tags = {
    Project       = "aws-instance-benchmarks"
    BenchmarkSuite = each.key
  }
}

# SSM Parameters for benchmark configuration
resource "aws_ssm_parameter" "benchmark_config" {
  name  = "/aws-instance-benchmarks/config"
  type  = "String"
  value = jsonencode({
    benchmark_suites   = var.benchmark_suites
    instance_families  = var.instance_families
    s3_bucket         = aws_s3_bucket.benchmark_results.id
    max_duration      = var.max_benchmark_duration
    enable_spot       = var.enable_spot_instances
  })

  description = "Configuration for AWS Instance Benchmarks"
  
  tags = {
    Project = "aws-instance-benchmarks"
  }
}

# Outputs
output "benchmark_results_bucket" {
  description = "S3 bucket for storing benchmark results"
  value       = aws_s3_bucket.benchmark_results.id
}

output "benchmark_vpc_id" {
  description = "VPC ID for benchmark environment"
  value       = aws_vpc.benchmark_vpc.id
}

output "benchmark_subnet_ids" {
  description = "Subnet IDs for benchmark instances"
  value       = aws_subnet.benchmark_public[*].id
}

output "benchmark_security_group_id" {
  description = "Security group ID for benchmark instances"
  value       = aws_security_group.benchmark_sg.id
}

output "benchmark_launch_template_id" {
  description = "Launch template ID for benchmark instances"
  value       = aws_launch_template.benchmark_template.id
}

output "benchmark_log_groups" {
  description = "CloudWatch log groups for benchmark suites"
  value       = { for k, v in aws_cloudwatch_log_group.benchmark_logs : k => v.name }
}