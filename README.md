dp-frontend-geography-controller
==================

An HTTP service for the controlling of the geography data relevant to a particular dataset.

### Configuration

| Environment variable | Default                 | Description
| -------------------- | ----------------------- | --------------------------------------
| BIND_ADDR            | :23700                  | The host and port to bind to.
| RENDERER_URL         | http://localhost:20010  | The URL of dp-frontend-renderer.
| ZEBEDEE_URL          | http://localhost:8082   | The URL of zebedee.
| SLACK_TOKEN          |                         | A slack token to write feedback to slack

### Licence

Copyright ©‎ 2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

