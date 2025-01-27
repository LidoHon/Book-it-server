# Use the official Golang image as the base for the build stage
FROM golang:1.20 AS build

# Set the working directory inside the container
WORKDIR /go/src/app

# Copy the Go application code into the container
COPY go-app/ .

# Install dependencies
RUN go mod tidy

# Build the Go application and output the binary to /go/bin/app
RUN go build -o /go/bin/app .

# Start a new stage for the final image
FROM alpine:latest

# Install necessary dependencies (e.g., ca-certificates)
RUN apk --no-cache add ca-certificates

# Copy the built Go binary from the build stage
COPY --from=build /go/bin/app /app

# Expose the port that the app will run on (use the correct port for your app)
EXPOSE 8000

# Command to run the app
CMD ["/app"]
