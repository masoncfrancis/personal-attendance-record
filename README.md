# personal-attendance-record (`par`)

I wrote this tool, called `par` for short, to automate keeping my own record of my in-office attendance. 

It will check if you are in office by means of seeing what you can reach on the network, assuming that your office network has different resources or is configured differently than your network at home. 

`par` is designed to be run periodically by a cronjob, and stores the record in a CSV file.

## Getting Started

### Configuration

All settings are set via environment variables. You can use a `.env` file kept in the same directory as the executable. An example is available in [`.env.example`](.env.example).

#### File Path

The file where the attendance record will be kept must be set in the `PAR_FILE` environment variable. It will be stored in CSV format so you can open it with Excel or another spreadsheet program. 

#### Modes

`par` can be configured to run in 2 modes: `corporateproxy` or `checkhost`. This mode is set via the `PAR_MODE` environment variable. 

##### `corporateproxy` mode

Use `corporateproxy` mode if you have to configure proxy settings on your machine in order to access the internet while on the office network. If your workplace network does not use a proxy through which your computer must route all internet traffic, this mode will **not** work for you. 

This mode works by checking if you are able to access the internet without going through a corporate proxy. 

###### How It Works:

The tool will attempt to query a URL. This URL can be set in the `PAR_URL` environment variable, and needs to be set to a website on the public internet that ordinarily would be unaccessible without using the corporate proxy. If `PAR_URL` is 
not set, it defaults to `https://www.google.com`. It will also send a ping request to the
proxy server set in the `PAR_PROXY_ADDRESS` environment variable. 

If the URL is not reachable directly but the ping to the proxy is successful, it assumes you are in the office. 
If it can reach the URL directly, it assumes you are working outside of the office.

##### `checkurl` mode

Use `checkurl` mode if you know there are resources that you cannot reach while off of the office network.

This mode works by checking if you can reach the URL configured in the `PAR_URL` environment variable. This variable should be set to a URL that is only reachable when on the office network.

###### How It Works:

`par` will try to reach the URL provided. If it is successful, it assumes you are in the office. If not, it assumes you are working from outside the office. 

### Command Line Usage

`par` exposes a simple command-line interface intended for use from a cronjob or manual invocation.

- `--check`, `-c`: Run a single attendance check, update log and exit
- `--path`, `-p`: Display the path of the file where the attendance record is kept
- `--version`, `-v`: Display the software version

### Crontab / Scheduling (recommended)

You can use a cron job to schedule `par` to run on a schedule. The instructions below are an example of how one might schedule `par` to log their attendance. 

To run `par` four times between 9:00 and 17:00 (9am–5pm), pick four times that suit you. A common choice is `09:30`, `11:30`, `15:00` (3:00 PM), and `16:30` (4:30 PM).

Create the cron jobs with `crontab -e` to run at those times and append output to a log. Because the minutes differ between entries in our example times, we will use two cron lines:

```
# Run at 09:30, 11:30, 15:00 and 16:30 every day
30 9,11,16 * * * /full/path/to/par --check >> /var/log/par-check.log 2>&1
0 15 * * * /full/path/to/par --check >> /var/log/par-check.log 2>&1
```

- **Use absolute paths:** cron runs with a minimal environment — always use the full path to the `par` executable.

## License

This code is provided under an MIT license. Use it how you like. I'm not responsible for how you use this code. If you rely on this software to keep track of your attendance, you do so at your own risk. I'm not liable for any consequences you experience related to this software.
