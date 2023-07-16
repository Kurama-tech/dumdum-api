# By Meer Sawood <msawood@redhat.com> 
# uses ubi minimal image to build and httpd as the server

# Use an official Red Hat UBI minimal runtime as a parent image
FROM registry.access.redhat.com/ubi8/nodejs-18:1-47.1687186189

# Set the working directory to /app
WORKDIR /app
USER 0


# Copy the current directory contents into the container at /app
COPY . /app

# Install dependencies
RUN npm install
RUN npm install -g pm2

RUN chown -R 1001:0 /app
USER 1001

#RUN npm run build

# Expose the server on port 80
EXPOSE 3000

# Start Server
# Set the command to run the app
ENTRYPOINT [ "/app/hack/entrypoint.sh" ]
CMD [ "sh" ]