# Use the official Golang image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /src

# Copy the local package files to the container's workspace
COPY src/geomelody .

# Download Go dependencies
RUN go get -d -v ./...

# Build the Go application
RUN go build -o main .

# Expose the port on which the microservice will run
EXPOSE 8080

# Run the microservice when the container starts
CMD ["./main"]
