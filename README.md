
# personal-attendance-record (`par`)

`par` is a tool to automate keeping your own record of in-office attendance. It determines if you are in the office by checking network reachability, assuming your office network has unique resources or configurations. Records are stored in a CSV file for easy access.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Installation/Updating](#installationupdating)
- [Building from Source](#building-from-source)
- [Configuration](#configuration)
- [Command Line Usage](#command-line-usage)
- [Scheduling](#scheduling)
- [FAQ](#faq)
- [License](#license)

---

## Getting Started


### Installation/Updating


#### UNIX-like Operating Systems (Linux, macOS, etc.)


Download the binary from the [Releases page](https://github.com/masoncfrancis/personal-attendance-record/releases) and save it in a folder of your choice (for example, `~/par-tools` or `/opt/par`).

```bash
mkdir -p ~/par-tools
mv path/to/downloaded/binary ~/par-tools/par
chmod +x ~/par-tools/par
```

You can run the program directly from that folder:

```bash
~/par-tools/par --version
```

Optionally, you may add the folder to your `PATH` if you want to run `par` manually from anywhere. For example, to add `~/par-tools` to your PATH, add the below line to your `~/.bashrc` or `~/.zshrc` file:

```bash
export PATH="$HOME/par-tools:$PATH"
```

Then, restart your shell.


##### If you run into problems running `par` after downloading/installing

If you get a `permission denied` error, you may need to run `chmod +x /path/to/par` to make the binary executable. On Macs, you may also need to run `xattr -d com.apple.quarantine /path/to/par` to remove the quarantine attribute that macOS applies to files downloaded from the internet.


#### Windows


Windows binaries are available, but not officially supported. Download the appropriate binary from the Releases page. You can run it from Command Prompt or PowerShell:

```powershell
cd path\to\downloaded\binary
./par.exe --version
```

If you encounter issues, please report them.



---

#### Building from Source


You'll need [Go](https://go.dev/dl/) (version 1.13 or later) installed.


Verify Go installation:

```bash
go version
```

Clone the repository:

```bash
git clone https://github.com/masoncfrancis/personal-attendance-record.git
cd personal-attendance-record
```


Then, build the binary (run `go mod tidy` if dependencies are missing):

```bash
go build -o par ./cmd/par
```


This creates an executable named `par` in the current directory. Move it to a directory in your `PATH` if desired.


---

### Configuration

All settings are set via environment variables. Before using `par`, you will need to configure these variables. You should create a `.env` file in the same directory as the executable. An example is available in [`.env.example`](.env.example).

#### Required Environment Variables

| Variable              | Description                                                      |
|-----------------------|------------------------------------------------------------------|
| `PAR_FILE`            | Path to the CSV file for attendance records                      |
| `PAR_MODE`            | Check method: `corporateproxy` or `checkurl`                    |
| `PAR_URL`             | URL to check (see method details below) (optional if using `corporateproxy` mode) |
| `PAR_PROXY_ADDRESS`   | Proxy address (only required for `corporateproxy` mode)          |
| `PAR_RECORD_INCREMENTS` | `daily` (default) or `all` (see below)                         |



Example `.env`:

```env
PAR_FILE=/path/to/attendance.csv
PAR_MODE=checkurl
PAR_RECORD_INCREMENTS=daily
PAR_URL=https://internal.company.com
PAR_PROXY_ADDRESS=http://proxy.company.com:8080
```


#### Record-Keeping Increments

Set `PAR_RECORD_INCREMENTS` to select how many records are saved:

| Value    | Description                                                                 |
|----------|-----------------------------------------------------------------------------|
| `daily`  | Only one entry per day, no matter how many checks are performed. If any check finds you in office, marks day as 'in'. |
| `all`    | Every check is logged with timestamp.                                       |

#### Check Method

Set `PAR_MODE` to select the check method:

| Value             | Description                                                                                 |
|-------------------|---------------------------------------------------------------------------------------------|
| `corporateproxy`  | Checks if you can access a URL without a proxy and pings the proxy server.                  |
| `checkurl`        | Checks if you can reach a URL only available on the office network.                         |

**corporateproxy example:**

- `PAR_URL=https://www.google.com` (default)
- `PAR_PROXY_ADDRESS=proxy.company.com:8080`

If the URL is not reachable directly but the proxy is reachable, you are assumed to be in the office.

**checkurl example:**

- `PAR_URL=https://internal.company.com`

If the URL is reachable, you are assumed to be in the office.



---

### Command Line Usage

`par` exposes a simple command-line interface for use in cronjobs or manual invocation.

| Flag       | Shorthand | Description                                         |
|------------|-----------|-----------------------------------------------------|
| `--check`  | `-c`      | Run a single attendance check, update log and exit  |
| `--no-save`| `-n`      | Run a check and print result (no file saved); good for testing |
| `--path`   | `-p`      | Display the path of the attendance record file      |
| `--version`| `-v`      | Display the software version                        |

Examples:

```bash
par --check
par -n
par --path
par -v
```



---

### Scheduling

#### UNIX-like (cron)

To run `par` periodically (e.g., 09:30, 11:30, 15:00, 16:30), add to your crontab (`crontab -e`):

```cron
# Run at 09:30, 11:30, 15:00 and 16:30 every day
30 9,11,16 * * * /full/path/to/par --check
0 15 * * * /full/path/to/par --check
```

**Note:** Always use the absolute path to the `par` executable.

#### Windows (Task Scheduler)

On Windows, use Task Scheduler to run `par.exe` at your desired times. Example action:

```
Program/script: C:\full\path\to\par.exe
Add arguments: --check
```


---

### FAQ


#### Why is my attendance saved in CSV format?

The attendance record is saved in CSV format mainly to make it easy to open with spreadsheet software like Excel or Google Sheets, as opposed to keeping it in a database or a more complex file format. This allows for easy viewing, editing, and analysis of your attendance data without needing specialized software.


#### Virus Warnings

Go binaries are structured differently than many other programs. Some antivirus software may flag the binaries as suspicious because they are not signed. If you encounter this, add an exception for the binary or build from source for peace of mind.

#### I found a bug

Please open an issue on the [GitHub Issues page](https://github.com/personal-attendance-record/issues)

#### I have a feature request

Please post in the [GitHub Discussions page](https://github.com/personal-attendance-record/discussions)


---

## License

This code is provided under the MIT license. Use it as you like. The author is not responsible for how you use this code or any consequences related to its use.
