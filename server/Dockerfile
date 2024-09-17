# Use the official Golang image as the base image
FROM golang:1.23

# Set the working directory inside the container
WORKDIR /usr/src/app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download & go mod verify

# Copy the rest of the application code
COPY . .

# Build the Go server binary
RUN go build -v -o /usr/local/bin/ ./...

# Start the Go server
CMD ["/usr/local/bin/delivery"]

