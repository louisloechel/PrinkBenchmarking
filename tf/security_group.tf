resource "aws_security_group" "my_app" {
  name   = "Ganges eval security group"
  vpc_id = aws_vpc.main.id
 
  ingress {
    cidr_blocks = [
      "0.0.0.0/0"
    ]
    from_port = 22
    to_port   = 22
    protocol  = "tcp"
  }
 
  ingress {
    cidr_blocks = [
      "0.0.0.0/0"
    ]
    from_port = 80
    to_port = 80
    protocol = "tcp"
  }

  // docker host
  ingress {
    cidr_blocks = ["10.0.0.0/16"]
    from_port = 2375
    to_port = 2375
    protocol = "tcp"
  }

  // benchmark ports
  ingress {
    cidr_blocks = ["10.0.0.0/16"]
    from_port = 50051
    to_port = 60000
    protocol = "tcp"
  }
 
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}