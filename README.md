***
⚠️ This repository will be archived in November 2023 as it is no longer in development. ⚠️
***

dp-frontend-geography-controller
==================

An HTTP service for connecting the geography datasets to the cmd datasets journey.

### Configuration

| Environment variable         | Default                 | Description
| ---------------------------- | ----------------------- | --------------------------------------
| BIND_ADDR                    | :23700                  | The host and port to bind to.
| RENDERER_URL                 | http://localhost:20010  | The URL of dp-frontend-renderer.
| CODELIST_API_URL             | http://localhost:22400  | The URL of the code list api.
| DATASET_API_URL              | http://localhost:22000  | The URL of the dataset api.
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                      | The graceful shutdown timeout in seconds
| HEALTHCHECK_INTERVAL         | 30s                     | The time between calling healthcheck endpoints for check subsystems
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                     | The time taken for the health changes from warning state to critical due to subsystem check failures

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details

### Licence

Copyright ©‎ 2018-2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

