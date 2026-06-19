[![Go Reference](https://pkg.go.dev/badge/github.com/cadence-workflow/starlark-worker.svg)](https://pkg.go.dev/github.com/cadence-workflow/starlark-worker)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fcadence-workflow%2Fstarlark-worker.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fcadence-workflow%2Fstarlark-worker?ref=badge_shield)

Starlark ⭐ meets Cadence ⚙️

Cadence Starlark Worker integrates Cadence Workflow with Starlark scripting language. It allows users to write workflows in Starlark, which are then run on Cadence without requiring worker redeployment. This approach enables flexible, multi-tenant workflow execution, combining Cadence's robustness with Starlark's simplicity and safety features.

## Developer Guide
This section is intended for contributors. Below are the instructions for setting up the development environment and running test workflows.

1. **Run Cadence via Docker Compose**:
   Follow the instructions in [this guide](https://github.com/cadence-workflow/cadence/tree/master/docker#quickstart-for-development-with-local-cadence-server) to start Cadence.

2. **Create the `default` domain**:
   Execute the following command:
   ```sh
   docker run -it --rm --network host ubercadence/cli:master --domain default domain register --retention 1
   ```
   You should see the domain settings page at [http://localhost:8088/domains/default/settings](http://localhost:8088/domains/default/settings).

3. **Start the Cadence Starlark Worker**:
   ```sh
   go run ./cmd/service 
   ```
   **Or start the temporal Starlark Worker**
   ```sh
   go run ./cmd/service --backend temporal --url localhost:7233 
   ```
4. **Run a test workflow**:
   ```sh
   go run ./cmd/cadence_client_main run --file ./test/testdata/ping.star
   ```
   ```sh
   go run ./cmd/temporal_client_main run --file ./test/testdata/ping.star
   ```

   Or try another test workflow which accepts custom input:
   ```sh
   go run ./client_main run --file ./testdata/concurrent_hello.star --function wf --args "[5, 1]"
   ```

5. **Modify and re-run the test workflow**:
   Make changes to the [testdata/ping.star](./testdata/ping.star) file and run the workflow again. The changes will take effect immediately without needing to restart the worker.

## Contributing

We'd love to have you contribute! Please see our [CONTRIBUTING.md](CONTRIBUTING.md) for more information. For general guidance on contributing to Cadence projects, check out our [Contributing Guide](https://github.com/cadence-workflow/cadence/blob/master/CONTRIBUTING.md).

Join us on [CNCF Slack](https://communityinviter.com/apps/cloud-native/cncf) to discuss and ask questions.


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fcadence-workflow%2Fstarlark-worker.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fcadence-workflow%2Fstarlark-worker?ref=badge_large)