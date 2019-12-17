
# Snapr CLI

Snaps a webcam image using `ffmpeg`.
Works on Linux and Mac computers.

## TODO

- Pidgeon command
- Test with PAM and Crontab

## Build

For your OS:
```
go build snapr
```

For specific OS:
```
GOOS=linux go build snapr
GOOS=darwin go build snapr
```

## Snap Command

Snaps a webcam photo:
```
snapr snap
snapr snap --help
snapr snap --capture-device /dev/video1
snapr snap --base-dir base/dir
snapr snap --extra-dir sub/dir
```

Example:
```
snapr snap -d /Users/joseph/Desktop
```

## Pidgeon Command

Uploads a photo to an AWS bucket:
```
snapr pideon
```