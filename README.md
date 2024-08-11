**uboa: A Local First HTTP Load Testing CLI Tool**
=============================================

uboa is a HTTP load testing tool designed to help you evaluate the performance and reliability of your web applications under various levels of concurrent traffic.

### Getting Started

To use uboa, simply run the command `uboa` followed by the required flags and options.

### Flags and Options

#### Required Flags

* `-u` or `--url`: The target URL to test (required)

#### Optional Flags

* `-H` or `--headers`: Custom HTTP headers (format: key1:value1,key2:value2)
* `-d` or `--body`: Request body for POST, PUT, or PATCH requests
* `-j` or `--json`: Output results in JSON format
* `--html` or `--html-output`: Output results in HTML format
* `-S` or `--skip-preview`: Skip automatic preview of results
* `-o` or `--output`: File path for saving the output (default: `{yyyy-mm-dd}_{method}_uboa-result`)
* `-c` or `--concurrency`: Number of concurrent clients (default: 5)
* `-n` or `--requests`: Total number of requests to send (default: 100)
* `-T` or `--timeout`: HTTP client timeout in seconds (default: 5)
* `-k` or `--keep-alive`: Enable HTTP keep-alive connections
* `-r` or `--max-retries`: Maximum allowed retry before erroring (default: 3)
* `-i` or `--insecure`: Use this flag to skip SSL verification
* `-ca` or `--ca`: CA certificate to verify peer against
* `-cert` or `--cert`: Client authentication certificate
* `-key` or `--key`: Client authentication key

### Example Usage

Here's an example of how to use uboa:
```bash
uboa -url https://google.com -method GET -concurrency 10 -requests 1000
```
This command will send 1000 GET requests to `https://google.com` using 10 concurrent clients.

### Installation  
Grab a pre-built binary from the [GitHub Releases page](https://github.com/oddegen/uboa/releases). You can optionally put the **uboa** binary in your `PATH` so you can run uboa from any location. Alternative you can build it from source,

To build uboa from source, you'll need to have Go installed on your system. Here are the steps:

1. Clone the repository: `git clone https://github.com/oddegen/uboa.git`
2. Change into the repository directory: `cd uboa`
3. Run `go build` to build the binary
4. Run `./uboa` to run the tool

### Output

uboa outputs the results of the load testing in a human-readable format. You can customize the output format using the `-json` or `-html` flags.
Here is an example of the rendered HTML output:

<img src="https://imgur.com/3oyxpH5" />
<img src="https://imgur.com/T9Ak0N1" />

### Contributing

uboa is an open-source project and welcomes contributions from the community. If you'd like to contribute, please refer to the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information.

### License

uboa is licensed under the MIT License. See the [LICENSE.md](LICENSE.md) file for more information.
