output "client_ip" {
  value = aws_instance.client.public_ip
}

output "server_ip" {
  value = aws_instance.server[*].public_ip
}

output "server_private_ip" {
  value = aws_instance.server[*].private_ip
}

output "client_private_ip" {
  value = aws_instance.client.private_ip
}