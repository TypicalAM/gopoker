############################
# STEP 1 build & run interactively using npm start
############################
FROM node:alpine

# Set up the workdir
WORKDIR /usr/src/gopoker
ENV NODE_ENV=development

# Copy the files
COPY package.json package-lock.json .

# Install the dependencies
RUN npm install

# Copy the rest of the files
COPY . .

# Start the app
CMD ["npm", "start"]
