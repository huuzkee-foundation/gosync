# GoSync

[![Build Status](https://drone.io/github.com/Nitecon/gosync/status.png)](https://drone.io/github.com/Nitecon/gosync/latest)

GoSync is a simple filesystem and directory replication system that uses a database for it's backend to track changes.
It was built out of necessity to auto scale web based properties where the sites were built as single server sites.

## Downloads:

  GoSync is built through drone.io so we will always have at least a linux 64bit binary available for download.  The 
  download includes the `gosync` binary a base configuration file and and an init.d script for redhat.

### Download location:
 Latest development binary is available here: [Latest Build](https://s3.amazonaws.com/nitecon-builds/gosync/latest/latest.tar.gz)
 
### Installation:
 Installation is very simple, download the latest build from the link above, extract the file it will contain 3 files:
 1. Extract the tarball, create a gosync config directory (/etc/gosync)
 2. Copy the configuration file into /etc/gosync... 
 3. Update the configuration file
 4. Copy the init script to /etc/init.d/ and make it executable with chmod +x 
 5. Edit the init.d/script and update the config file directive *if you did not copy the config file to `/etc/gosync/config.cfg`*
 6. Start the application by running it as root user with `/etc/init.d/gosync start`
 
### Configuration:
 Please check [Configuration](https://github.com/Nitecon/gosync/wiki/Configuration) page on the wiki

## Notice:
 GoSync is still under heavy development and there is a lot of stuff changing frequently, once it becomes stable there
 will be a tag added to indicate stable versions in order to track against those.  For now please make sure to check
 the bottom of this page for possible breaking changes and issues.
 
## Recent changes (may include breaking changes)
The configuration file was updated to reflect new changes, when downloading the latest build please make sure to
check and validate that your configuration file is updated with what has changed in the latest build.

New Changes Include:
- Updated log parameters to specify the log file
- Updated the log level params as strings to specify different levels (debug / info / error)
- Log levels in the config specify the lowest level that will be logged
- Log file location was slightly updated and stdout still works.