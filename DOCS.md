
Use this plugin to build and push Docker images to the Google Container Registry (GCR). Please read the GCR [documentation](https://cloud.google.com/container-registry/) before you begin. You will need to generate a [JSON token](https://developers.google.com/console/help/new/#serviceaccounts) to authenticate to the registry and push images.

The following parameters are used to configure this plugin:

* `debug` - enable debug mode.
* `dry-run` - enable dry-run, push is skipped.
* `registry` - authenticates to this registry (defaults to `gcr.io`).
* `auth_key` - json authentication key for service account.
* `storage_driver` - use `aufs`, `devicemapper`, `btrfs` or `overlay` driver.
* `repo` - repository name for the image.
* `tag` - repository tag for the image (defaults to `latest`).
* `dockerfile` - user custom Dockerfile (defaults to `Dockerfile`).
* `context` - docker context dir (defauults to `.`).
* `args` - docker build args.



Sample configuration:

```yaml
publish:
  image: foo/drone-gcr
  repo: foo/bar
  auth_key: >
    {
      "private_key_id": "...",
      "private_key": "...",
      "client_email": "...",
      "client_id": "...",
      "type": "..."
    }
```

Sample configuration using multiple tags:

```yaml
publish:
  image: foo/drone-gcr
  repo: foo/bar
  tag:
    - latest
    - "1.0.1"
    - "1.0"
  auth_key: >
    {
      "private_key_id": "...",
      "private_key": "...",
      "client_email": "...",
      "client_id": "...",
      "type": "..."
    }
```

## JSON auth key.

When setting your token in the `.drone.yml` file you must use [folded block scalars](http://www.yaml.org/spec/1.2/spec.html#id2796251) to avoid parsing errors:

```yaml
publish:
  auth_key: >
    {
      "private_key_id": "...",
      "private_key": "...",
      "client_email": "...",
      "client_id": "...",
      "type": "..."
    }
```

When injecting secrets you must also use a folded block scalar:
```yaml
publish:
  auth_key: >
    ${GCR_AUTH_KEY}
```
