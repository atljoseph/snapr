
# Snapr CLI

Snaps a webcam image using `ffmpeg`.
Works on Linux and Mac computers.

## TODO

- Upload command improvement - look up files from the base directory to upload
- Test with PAM and Crontab
- Add specific filename to snap command

## Environment

You need an environment file named `.env`:
```
S3_BUCKET=my.s3.bucket
S3_REGION=us-east-west
S3_TOKEN=ABC123
S3_SECRET=123ABC
```

To use a different file, like `.prod.env`, see the `Build` section (below).

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

To use a custom `.env` file:
```
go build -ldflags "-X main.EnvFilePath=.prod.env" snapr
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
snapr snap
```

## Upload Command

Uploads a photo to an AWS bucket:
```
snapr upload --base-dir my/base/dir
snapr upload --in-file my/in/file
```

Example:
```
snapr upload --in-file file.jpg
```