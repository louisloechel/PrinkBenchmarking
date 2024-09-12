# Use the official Golang image as the base image
FROM golang:1.21.5

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download


RUN curl -fsSL https://get.docker.com -o get-docker.sh
RUN sh get-docker.sh

# Copy the client directory to the Working Directory inside the container
COPY ./ .

# Build the Go app
RUN go build -o client ./main.go


# Command to run the executable
CMD ["./client"]