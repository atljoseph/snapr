
# Snapr CLI

Snaps a webcam image using `ffmpeg`.
Works on Linux and Mac computers.

## TODO

- Pidgeon command improvement - look up files from the base directory
- Test with PAM and Crontab

## Environment

You need an environment file named `.env`:
```
S3_BUCKET=my.s3.bucket
S3_REGION=us-east-west
S3_TOKEN=ABC123
S3_SECRET=123ABC
```

## Build

Upon building, the environment is compiled into the binary.
After building, you will no longer need the `.env` file.

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
snapr snap --base-dir my/base/dir
snapr snap --extra-dir my/sub/dir
```

Example:
```
snapr snap --base-dir /Users/joseph/Desktop
```

## Pidgeon Command

Uploads a photo to an AWS bucket:
```
snapr pidgeon --base-dir my/base/dir
snapr pidgeon --in-file my/in/file
```