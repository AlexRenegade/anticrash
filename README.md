# AntiCrash

Simple bot designed to detect and delete files which crash the Discord client

### A hosted version is available to invite [here](https://discord.com/oauth2/authorize?client_id=839625900860899368&permissions=93184&scope=bot).

## Building

Ensure you have [go installed](https://golang.org/doc/install) before continuing.

You can build the application using the following command:

```
$ go build
```

This will create an executable (`AntiCrash` on Unix systems and `AntiCrash.exe` on Windows) you can use to run the application.

## Configuration

Create a `config.toml` file in the same directory as the executable. Enter the following lines:

```
Prefix = "+"
Token = "yourTokenHere"
ReplyToMessage = true
LogChannel = "000000000000000000"
```

If you set `ReplyToMessage` to `true`, the bot will send an inline reply to the message containing a crash file.

If you provide a value for `LogChannel`, the bot will log the user who posted the crash file, the URL to the file, and the channel it was posted in to the channel specified.

## Usage

The application uses `ffprobe` (from the ffmpeg suite) to parse files for the properties which cause crashes, so it needs to be installed on the system. 
ffprobe version 4 is what this application expects, however there is a possiblity that other versions will work, but it has not been tested.

Ubuntu 20.04 uses ffprobe as the default version, so it can simply be installed with:

```
$ apt-get install ffmpeg
```

On other operating systems, you may need to install ffprobe yourself. Make sure you add the location of ffprobe to PATH, so that AntiCrash can find it.

To run the application, simply run `./AntiCrash` on Unix, and `./AntiCrash.exe` on Windows. 

If you want to place the configuration file in a different location, you can specify the path to it as an argument: 

```
$ ./AntiCrash /path/to/config.toml
```