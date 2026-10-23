# personal-attendance-record (`par`)

I wrote this tool, called `par` for short, to automate keeping my own record of my in-office attendance. 

It will check if you are in office by means of seeing what you can reach on the network, assuming that your office network has resources or is configured differently than your network at home. 

`par` is designed to be run periodically as a cronjob.

## Getting Started

### Configuration

All settings are set via a `.env` file kept in the same directory as the executable. An example is available in [`.env.example`](.env.example).

#### Modes

`par` can be configured to run in 3 modes: `corporateproxy` and `checkhost`. This mode is set via the `PAR_MODE` environment variable. 

##### `corporateproxy` mode

Use `corporateproxy` mode if you have to configure proxy settings on your machine in order to access the internet while on the office network. If your workplace network does not use a proxy through which your computer must route all internet traffic, this mode will **not** work for you. 

This mode works by checking if you are able to access the internet without going through a corporate proxy. 

###### Logic:

The tool will attempt to query a URL. This URL can be set in the `CHECK_URL` environment variable, and needs to be set to a website on the public internet that ordinarily would be unaccessible when not using the corporate proxy. If `CHECK_URL` is 
not set, it defaults to `https://www.google.com`.

If the URL is not reachable directly, it assumes you are in the office. 
If it can reach the URL directly, it assumes you are working outside of the office.

##### `checkurl` mode

Use `checkurl` mode if you know there are resources that you cannot reach while off of the office network.

This mode works by checking if you can reach a URL, configured in the `CHECK_URL` environment variable. This should be set to a URL that is only  

###### Logic:

`par` will try to reach the URL provided. If it is successful, it assumes you are in the office. If not, it assumes you are working from outside the office. 


## License

This code is provided under an MIT license. Use it how you like. I'm not responsible for how you use this code. If you rely on this software to keep track of your attendance, you do so at your own risk. I'm not liable for any consequences you experience related to this software.
