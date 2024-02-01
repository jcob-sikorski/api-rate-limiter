# Pin specific version for stability
FROM nginx:1.25.3-alpine AS base

# Specify working directory other than /
WORKDIR /etc/nginx

# Remove default configuration
RUN rm conf.d/default.conf

# Copy the Nginx configuration file
# This file configures Nginx to run as a reverse proxy
COPY nginx.conf /etc/nginx/nginx.conf

# Indicate expected port
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
