# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.22.2

# Create a directory inside the container to store all our application and then make it the working directory.
RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Build the Go app
RUN go build -o main .


# Run the executable
CMD ["./main"]