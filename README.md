# personal-attendance-record (`par`)

I wrote this tool, called `par` for short, to automate keeping my own record of my in-office attendance. 

It will check if you are in office by means of seeing what you can reach on the network, assuming that your office network has different resources or is configured differently than your network at home. 

`par` is designed to be run periodically by a cronjob, and stores the record in a CSV file.

## Getting Started

### Installation/Updating

#### UNIX-like Operating Systems (Linux, macOS, etc.)

Until I can get a package manager or install script set up, you'll need to download the binary manually from the [Releases page](https://github.com/masoncfrancis/personal-attendance-record/releases). Then, if you like, move it to a directory in your PATH. If you don't know how to do that, Google and ChatGPT are your friends.

##### If you run into problems running `par` after downloading/installing

If you get a `permission denied` error, you may need to run `chmod +x /path/to/par` to make the binary executable. On Macs, you may also need to run `xattr -d com.apple.quarantine /path/to/par` to remove the quarantine attribute that macOS applies to files downloaded from the internet.

#### Windows

While Windows binaries are available, I haven't tested them and they are not officially supported. If you are feeling adventurous, you can try downloading the appropriate binary for your architecture from the Releases page and using it as you please. Let me know if you run into any issues.


#### Building from source

You'll need to have [Go](https://go.dev/dl/) (version 1.13 or later) installed to build from source.

Once you have Go installed, you can build `par` from source. First, clone the repository:

```bash
git clone https://github.com/masoncfrancis/personal-attendance-record.git
cd personal-attendance-record
```

Then, build the binary:

```bash
go build -o par ./cmd/par
```

This will create an executable named `par` in the current directory. You can move it to a directory in your PATH if you like.

### Configuration

All settings are set via environment variables. You can use a `.env` file kept in the same directory as the executable. An example is available in [`.env.example`](.env.example).

#### File Path

The file where the attendance record will be kept must be set in the `PAR_FILE` environment variable. It will be stored in CSV format so you can open it with Excel or another spreadsheet program. 

#### Record-keeping increments

Users may have personal preferences over the granularity of records kept. As such, the user can choose whether to maintain 1 attendance entry per day or to save all checks. 
This is configured by setting the environment variable `PAR_RECORD_INCREMENTS` to `daily` or `all`, respectively.

##### `daily`

If the user elects to maintain 1 entry per day, `par` will continue to check as many times as it is run according to its established schedule, but it will only maintain 1 entry for each day. 

As such, if `par` finds that the user is in the office during at least 1 of its scheduled checks, it will mark the user as in-office for that day. The user must be working away from the office for all checks for `par` to record the day as not-in-office.

##### `all`

If the user elects to keep all records, `par` will save the result of all checks along with their timestamps in the file. 

#### Check method

The user can select a method by which `par` performs its checks: `corporateproxy` or `checkhost`. The method must be set via the `PAR_MODE` environment variable. 

##### `corporateproxy` method

Use the `corporateproxy` method if you have to use a corporate proxy to access the internet from the corporate network. If your computer does not require a proxy configuration to enable internet access while in the office, this mode will **not** work for you. If you use a corporate proxy to gain internet access even while working away from the office, this method also won't work for you. 

This method works by checking if you are able to access the internet without going through a corporate proxy. 

###### How the `corporateproxy` method works:

The tool will attempt to query a URL. This URL can be set in the `PAR_URL` environment variable, and needs to be set to a website on the public internet that ordinarily would be unaccessible without using the corporate proxy. If `PAR_URL` is 
not set, it defaults to `https://www.google.com`. It will also send a ping request to the
proxy server set in the `PAR_PROXY_ADDRESS` environment variable. 

If the URL is not reachable directly but the ping to the proxy is successful, it assumes you are in the office. 
If it can reach the URL directly, it assumes you are working outside of the office.

##### `checkurl` method

Use the `checkurl` method if you know there are network resources that you cannot reach while away from the office network.

This mode works by checking if you can reach the URL configured in the `PAR_URL` environment variable. This variable should be set to a URL that is only reachable when on the office network.

###### How the `checkurl` method works:

`par` will try to reach the URL provided. If it is successful, it assumes you are in the office. If not, it assumes you are working away from the office. 


### Command Line Usage

`par` exposes a simple command-line interface intended for use from a cronjob or manual invocation.

- `--check`, `-c`: Run a single attendance check, update log and exit
- `--no-save`, `-n`: Run a single attendance check and print the result (no file saved)
- `--path`, `-p`: Display the path of the file where the attendance record is kept
- `--version`, `-v`: Display the software version


### Crontab / Scheduling (recommended)

There are multiple ways to schedule software to run, but we recommend setting up a cron job. 

You can use a cron job to schedule `par` to run on a schedule. The instructions below are an example of how one might schedule `par` to log their attendance. 

To run `par` four times between 9:00 and 17:00 (9am–5pm), pick four times that suit you. A common choice is `09:30`, `11:30`, `15:00` (3:00 PM), and `16:30` (4:30 PM).

Create the cron jobs with `crontab -e` to run at those times and append output to a log. Because the minutes differ between entries in our example times, we will use two cron lines. paste the following into the file:

```
# Run at 09:30, 11:30, 15:00 and 16:30 every day
30 9,11,16 * * * /full/path/to/par --check
0 15 * * * /full/path/to/par --check
```

- **Use absolute paths:** cron runs with a minimal environment — always use the full path to the `par` executable.

### FAQ

#### Virus warning

Go programs are stuctured differently than a lot of other programs out there. Some antivirus software may flag the binaries as suspicious because of this, and that they are not signed. If you run into this issue, you can try adding an exception for the binary in your antivirus software. Or, if you're worried, feel free to build the program from source yourself. 


## License

This code is provided under an MIT license. Use it how you like. I'm not responsible for how you use this code. If you rely on this software to keep track of your attendance, you do so at your own risk. I'm not liable for any consequences you experience related to this software.
