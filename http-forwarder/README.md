# HTTP forwarder

HTTP forwarder provides message repository implementation for forwarder messages by HTTP.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                          | Description                                              | Default                |
|-----------------------------------|----------------------------------------------------------|------------------------|
| MF_NATS_URL                       | NATS instance URL                                        | nats://localhost:4222  |
| MF_HTTP_FORWARDER_LOG_LEVEL       | Log level for HTTP forwarder (debug, info, warn, error)  | error                  |
| MF_HTTP_FORWARDER_PORT            | Service HTTP port                                        | 8990                   |
| MF_HTTP_FORWARDER_REMOTE_URL      | Receiver of messages URL                                 | http://localhost:9000  |
| MF_HTTP_FORWARDER_REMOTE_TOKEN    | Receiver authorization bearer token                      | ""                     |
| MF_HTTP_FORWARDER_SUBJECTS_CONFIG | Configuration file path with subjects list               | /config/subjects.toml  |
| MF_HTTP_FORWARDER_CONTENT_TYPE    | Message payload Content Type                             | application/senml+json |

## Deployment

```yaml
  version: "3.7"
  http-forwarder:
    image: jonathandreyer/mainflux-http-forwarder:[version]
    container_name: [instance name]
    expose:
      - [Service HTTP port]
    restart: on-failure
    environment:
      MF_NATS_URL: [NATS instance URL]
      MF_HTTP_FORWARDER_LOG_LEVEL: [HTTP forwarder log level]
      MF_HTTP_FORWARDER_PORT: [Service HTTP port]
      MF_HTTP_FORWARDER_REMOTE_URL: [Receiver of messages URL]
      MF_HTTP_FORWARDER_REMOTE_TOKEN: [Receiver authorization bearer token]
      MF_HTTP_FORWARDER_SUBJECTS_CONFIG: [Configuration file path with subjects list]
      MF_HTTP_FORWARDER_CONTENT_TYPE: [Message payload Content Type]
    ports:
      - [host machine port]:[configured HTTP port]
    volumes:
      - ./subjects.toml:/config/subjects.toml
```

To start the service, execute the following shell script:

```bash
# download the latest version of the service
git clone https://github.com/jonathandreyer/mainflux-http-forwarder

cd mainflux

# compile the http-forwarder
make

# copy binary to bin
make install

# Set the environment variables and run the service
MF_NATS_URL=[NATS instance URL] MF_HTTP_FORWARDER_LOG_LEVEL=[HTTP forwarder log level] MF_HTTP_FORWARDER_PORT=[Service HTTP port] MF_HTTP_FORWARDER_REMOTE_URL=[Receiver of messages URL] MF_HTTP_FORWARDER_REMOTE_TOKEN=[Receiver authorization bearer token] MF_HTTP_FORWARDER_SUBJECTS_CONFIG=[Configuration file path with subjects list] MF_HTTP_FORWARDER_CONTENT_TYPE=[Message payload Content Type]
```

### Using docker-compose

This service can be deployed using docker containers.
Docker compose file is available in `<project_root>/docker/addon/http-forwarder/docker-compose.yml`.
In order to run Mainflux HTTP forwarder, execute the following command:

```bash
docker-compose -f docker/addon/http-forwarder/docker-compose.yml up -d
```

_Please note that you need to start core services before the additional ones._

## Usage

Starting service will start consuming normalized messages in SenML format.

[doc]: http://mainflux.readthedocs.io
