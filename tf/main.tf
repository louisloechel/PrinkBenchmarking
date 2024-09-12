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
  instance_type               = var.instance_type_client
  availability_zone           = var.availability_zone
  security_groups             = [aws_security_group.my_app.id]
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.my_app.id

  root_block_device {
    volume_size = 50
  }
 
  key_name = "ssh-key"
 
  ### Install Docker
  user_data = <<-EOF
  #!/bin/bash
  sudo yum update -y
  sudo yum install -y docker git
  sudo service docker start
  sudo usermod -a -G docker ec2-user
  sudo curl -L https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m) -o /usr/local/bin/docker-compose
  sudo chmod +x /usr/local/bin/docker-compose

  git clone https://github.com/louisloechel/PrinkBenchmarking.git /home/ec2-user/PrinkBenchmarking
  chown ec2-user:ec2-user /home/ec2-user/PrinkBenchmarking -R


  EOF
 
  tags = {
    Name = "my_app_CLIENT"
  }
}


resource "aws_instance" "server" {
  ami                         = var.instance_ami
  instance_type               = var.instance_type_server
  availability_zone           = var.availability_zone
  security_groups             = [aws_security_group.my_app.id]
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.my_app.id

  count = 10
 
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
    Name = "my_app_SERVER-${count.index}"
  }
}