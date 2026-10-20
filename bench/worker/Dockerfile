FROM node:22-bookworm-slim
WORKDIR /w
RUN apt-get update -qq && apt-get install -y -qq ca-certificates >/dev/null && npm init -y >/dev/null && npm i workerd >/dev/null 2>&1
COPY config.capnp ./
COPY src ./src
EXPOSE 8787
CMD ["npx","workerd","serve","config.capnp"]
