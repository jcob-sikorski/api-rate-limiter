# Pin specific version for stability
# Use slim for reduced image size
FROM node:21.6-alpine3.18 AS base

# Set NODE_ENV
ENV NODE_ENV production

# Specify working directory other than /
WORKDIR /usr/src/app

# Copy only files required to install
# dependencies (better layer caching)
COPY package*.json ./

RUN --mount=type=cache,target=/usr/src/app/.npm \
  npm set cache /usr/src/app/.npm && \
  npm install

# Use non-root user
# Use --chown on COPY commands to set file permissions
USER node

# Copy remaining source code AFTER installing dependencies. 
# Again, copy only the necessary files
COPY --chown=node:node . .

# Indicate expected port
EXPOSE 3000

CMD [ "node", "index.mjs" ]