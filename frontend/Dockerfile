############################
# STEP 1 build
############################
FROM node:14-alpine AS builder

ARG API_URL
ENV REACT_APP_API_URL $API_URL
ENV NODE_ENV production

# Add a work directory
WORKDIR /app

# Cache and Install dependencies
COPY package.json .
RUN npm install --omit=dev

# Copy app files
COPY . .

# Build the app
RUN npm run build

# Bundle static assets with nginx
FROM nginx:1.21.0-alpine as production
ENV NODE_ENV production

# Copy built assets from builder
COPY --from=builder /app/build /usr/share/nginx/html

# Add your nginx.conf
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
