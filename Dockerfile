# Use the official Go image as the base
FROM python:3.11

# Set the working directory
WORKDIR /app

# Install Go
ENV GO_VERSION=1.21.0
RUN wget -q https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

# Install requirments
COPY requirements.txt .
RUN pip install -r requirements.txt

# run serve.go
ENV PATH=$PATH:/usr/local/go/bin
CMD ["go", "run", "serve.go"]