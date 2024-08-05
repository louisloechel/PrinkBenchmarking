# Use the official Golang image as the base image
FROM golang:1.19

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o client ./main.go

# Expose port 50051 and 50052 to the outside world
EXPOSE 50051
EXPOSE 50052

# Command to run the executable
CMD ["./client"]