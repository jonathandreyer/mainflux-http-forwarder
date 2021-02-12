# Mainflux HTTP forwarder

This repository is an add-on to [Mainflux](https://github.com/mainflux/mainflux) project.
It allows to forward messages to a specific endpoint by HTTP protocol. Messages can be 
filtered by subjects as explained [here](https://mainflux.readthedocs.io/en/latest/messaging/#subtopics) 
by the [configuration file](docker/addons/subjects.toml).

## Features
- Forwards NATS messages by HTTP

## License

[Apache-2.0](LICENSE)
