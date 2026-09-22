FROM node:24-alpine AS builder
WORKDIR /src
COPY package*.json ./
RUN npm ci
COPY tsconfig.json ./
COPY src ./src
RUN npm run build

FROM node:24-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
COPY --from=builder /src/dist ./dist
EXPOSE 3001
CMD ["npm", "start"]