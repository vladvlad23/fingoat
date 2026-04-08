FROM node:22-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json .
RUN npm ci --omit=dev
EXPOSE 3000
ENV HOST=0.0.0.0
ENV PORT=3000
CMD ["node", "build"]
