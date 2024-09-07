provider "aws" {
  region = var.aws_region
  shared_credentials_files = ["~/.aws/credentials"]
  profile = "default"
}


resource "aws_key_pair" "ssh-key" {
  key_name   = "ssh-key"
  public_key = var.ssh_public_key
}


resource "aws_instance" "client" {
  ami                         = var.instance_ami
  instance_type               = var.instance_type
  availability_zone           = var.availability_zone
  security_groups             = [aws_security_group.my_app.id]
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.my_app.id
 
  key_name = "ssh-key"
 
  ### Install Docker
  user_data = <<-EOF
  #!/bin/bash
  sudo yum update -y
  sudo yum install -y docker
  sudo service docker start
  sudo usermod -a -G docker ec2-user
  EOF
 
  tags = {
    Name = "my_app_CLIENT"
  }
}


resource "aws_instance" "server" {
  ami                         = var.instance_ami
  instance_type               = var.instance_type
  availability_zone           = var.availability_zone
  security_groups             = [aws_security_group.my_app.id]
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.my_app.id
 
  key_name = "ssh-key"
 
  ### Install Docker
  user_data = <<-EOF
  #!/bin/bash
  sudo yum update -y
  sudo yum install -y docker

  sudo mkdir -p /lib/systemd/system/docker.service.d
  cat << EFF > /lib/systemd/system/docker.service.d/override.conf
  [Service]
  ExecStart=
  ExecStart=/usr/bin/dockerd -H fd:// -H tcp://0.0.0.0:2375
  EFF

  sudo systemctl daemon-reload
  sudo service docker restart
  sudo usermod -a -G docker ec2-user

  EOF
 
  tags = {
    Name = "my_app_SERVER"
  }
}