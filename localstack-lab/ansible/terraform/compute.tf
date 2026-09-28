# 1. VPC (Virtual Private Cloud) Katmanı
resource "aws_vpc" "lab_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-vpc"
  })
}

# 2. Subnet Katmanı
resource "aws_subnet" "lab_subnet" {
  vpc_id                  = aws_vpc.lab_vpc.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-subnet"
  })
}

# 3. Security Group (Güvenlik Duvarı)
resource "aws_security_group" "lab_sg" {
  name        = "${local.name_prefix}-web-sg"
  description = "Allow SSH and HTTP access"
  vpc_id      = aws_vpc.lab_vpc.id

  ingress {
    description = "SSH Access"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTP Access"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.tags
}

# 4. SSH Key Pair (Sunucuya Erişim Anahtarı)
resource "aws_key_pair" "lab_key" {
  key_name   = "${local.name_prefix}-key"
  public_key = var.ssh_public_key

  tags = local.tags
}

# 5. EC2 Instance (Sanal Sunucu Katmanı)
resource "aws_instance" "web_server" {
  ami                    = "ami-00000000" # LocalStack varsayılan mock AMI
  instance_type          = "t2.micro"
  subnet_id              = aws_subnet.lab_subnet.id
  vpc_security_group_ids = [aws_security_group.lab_sg.id]
  key_name               = aws_key_pair.lab_key.key_name

  user_data = <<-EOF
              #!/bin/bash
              echo "Hello from LocalStack EC2!" > index.html
              python3 -m http.server 80 &
              EOF

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-ec2-web"
  })
}

# 6. S3 Obje Yükleme (S3 Katmanında Dosya Yönetimi Pratiği)
resource "aws_s3_object" "sample_file" {
  bucket       = module.demo_bucket.bucket
  key          = "welcome.txt"
  content      = "Hello Hıdırcan! S3 object created via Terraform in LocalStack."
  content_type = "text/plain"

  tags = local.tags
}
