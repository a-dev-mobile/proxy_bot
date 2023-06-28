# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:latest

# Create a directory inside the container to store all our application and then make it the working directory.
RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Build the Go app
RUN go build -o main .

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["./main"] >> /var/log/app.log 2>&1