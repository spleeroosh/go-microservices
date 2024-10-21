FROM alpine:latest

# Create an application directory
RUN mkdir /app

# Set the working directory to /app
WORKDIR /app

# Copy the binary to /app
COPY mailerApp /app

# Copy the templates directory into the container
COPY templates /app/templates

# Set the command to run your application
CMD [ "/app/mailerApp" ]