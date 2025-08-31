# Middlewarr

Middlewarr is a lightweight proxy for \*arr applications.

## Installation

1. Copy the sample configuration file:

   ```bash
   cp settings.sample.yml /path/to/middlewarr/data/settings.yml
   ```

2. Run the container using Docker:

   ```bash
   docker run -p 9292:80 \
     -v /path/to/middlewarr/data:/data \
     --restart unless-stopped \
     ghcr.io/middlewarr/middlewarr:latest
   ```

### Configuration

The minimum configuration required is a settings.yml file placed in the /data directory:

```yaml
apiKey: example
host: 0.0.0.0

templates:
  repository: https://github.com/middlewarr/templates
  branch: main

log:
  level: 1
```

### Templates

Templates are loaded from the configured repository at server startup.
They define endpoint configurations for each supported service.
