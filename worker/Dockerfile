FROM alpine:latest

# Install necessary packages
RUN apk add --no-cache dcron redis nano

# Copy your worker into the container
COPY worker.sh /worker.sh

# Make your worker executable
RUN chmod +x /worker.sh

# Add a new cron job to run your worker every minute
RUN echo "* * * * * /worker.sh" | crontab -

# Run the cron daemon in the foreground
CMD ["crond", "-f"]
