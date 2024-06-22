# End-to-end Testing

End-to-end tests for Terralist. This package uses [venom](https://github.com/ovh/venom) to simplify the E2E suites.

## Testing base image

We build a container image with all the dependencies needed to run the E2E tests. You can check its page here: [ghcr.io/terralist/test-base](https://github.com/terralist/terralist/pkgs/container/test-base).

You can spin a one-off container using the docker-cli (or alternatives):
```bash
docker run --rm -v "${PWD}/:/app/" --it ghcr.io/terralist/test-base:latest bash
```

NOTE: Assuming your current working directory is the repository base.

## Launching a Terralist test server

To make the tests run, you need a server to test against. You can build Terralist from the source code using the build task (directly in the test container):
```bash
task build -- release
nohup ./terralist server >/dev/null 2>&1 &
```

or, you can spin a new container with the official Terralist image:
```bash
docker run --rm -e ... -p 5758:5758 --it ghcr.io/terralist/terralist:latest server
```

Of course, you will need to configure Terralist accordingly in both cases.

## Running the suites

Depending on your method of choice for deploying the test server, you might need to modify the `url` attribute from the `./e2e/variables.yaml` file.

When everything is set and ready, you can launch the test execution using the following venom-cli command:
```bash
venom run --var-from-file ./e2e/variables.yaml ./e2e/suites
```
