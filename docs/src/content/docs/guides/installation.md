---
title: Installation
description: Guide on how to install Koito to start tracking your listening history.
---

## Docker
By far the easiest way to get up and running with Koito is using docker. Here is an example Docker Compose file to get you up and running in minutes:
```yaml title="compose.yaml"
services:
  koito:
    image: gabehf/koito:latest
    container_name: koito
    environment:
      - KOITO_SQLITE_ENABLED=true # will not be needed after v0.2.0
    ports:
      - "4110:4110"
    volumes:
      - ./koito:/etc/koito
    restart: unless-stopped

```

`KOITO_SQLITE_ENABLED=true` is highly recommended for new installations, as PostgreSQL support will be deprecated in v0.2.0 onwards. After v0.2.0, you will not need `KOITO_SQLITE_ENABLED` anymore.

You can find a full list of configuration options in the [configuration reference](/reference/configuration).

## Build from source

If you don't want to use docker, you can also build the application from source.

First, you need to install dependencies. Koito relies on `make`, `yarn` for building the client, and `libvips-dev` to process images.

```sh
sudo apt install libvips-dev make npm
sudo npm install --global yarn
```

If you aren't installing on an Ubuntu or Debian based system, you can easily find ways to install make, npm, and yarn by googling, and you can find other ways to install `libvips-dev` on the [libvips wiki](https://github.com/libvips/libvips/wiki/).

Then, clone the repository and execute the build command using the included Makefile:

```sh
git clone https://github.com/gabehf/koito && cd koito
make build
```

When the build is finished, you can run the executable at the root of the directory.

```sh
KOITO_SQLITE_ENABLED=true ./koito
```

Then, navigate your browser to `localhost:4110` to enter your Koito instance.
