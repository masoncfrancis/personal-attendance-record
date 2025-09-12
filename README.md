# personal-attendance-logger

This is a Python script that will check if you are in office and save its finding to a CSV file.

It works by checking if you are able to access the internet without going through a corporate proxy. If a proxy
is required, it assumes you are in the office and on the corporate network. If not, it assumes you are working
remotely and can access the internet directly. 

It is intended to be used as a cronjob run multiple times per day. 
