# Use the official Go image
FROM golang:1.20

# Install Air
# RUN go install github.com/air-verse/air@latest

# Set the working directory
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Expose the port your application runs on
EXPOSE 8000

# Run Air as the default command
CMD ["/app/main"]