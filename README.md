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

3. **Start the Starlark Worker**:
   ```sh
   go run .
   ```

4. **Run a test workflow**:
   ```sh
   go run ./client_main run --file ./testdata/ping.star
   ```

5. **Modify and re-run the test workflow**:
   Make changes to the [testdata/ping.star](./testdata/ping.star) file and run the workflow again. The changes will take effect immediately without needing to restart the worker.
