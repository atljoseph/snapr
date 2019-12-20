
# Snapr CLI

Snaps a webcam image using `ffmpeg`.
Works on Linux and Mac computers.

## TODO

- Improve Upload command (dir walking) and add new tests
- Add Users to Snap Command testing and retreat the prepend as new dir
- Add List Command and tests
- Add Download Command and tests
- Test with PAM and Crontab (exit code 0 for pam)

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

## Environment

You need a file at the project root named `.env`, and yes, it can be blank.
For the environment variables specified in the `.env` file, after building, they will no longer be overridable at runtime. 
If a variable is not set, then it can be overriden at runtime.

These OPTIONAL variables apply to the `snap` command flags:
```
SNAP_DEVICE=
SNAP_DIR_EXTRA=
SNAP_DIR=
SNAP_FILE=
SNAP_FILE_FORMAT=
SNAP_FILE_USERS=
SNAP_UPLOAD_AFTER_SUCCESS=
SNAP_CLEANUP_AFTER_UPLOAD=
```

These OPTIONAL variables apply to the `upload` command flags:
```
UPLOAD_DIR=
UPLOAD_FILE=
UPLOAD_CLEANUP_AFTER_SUCCESS=
UPLOAD_FORMATS=
UPLOAD_LIMIT=
```

These REQUIRED variables are used for AWS:
```
S3_BUCKET=my.s3.bucket
S3_REGION=us-east-west
S3_TOKEN=ABC123
S3_SECRET=123ABC
```

For boolean inputs, the following inputs count as true:
```
1
true
TRUE
True
``` 

To use a different file, like `.prod.env`, see the `Build` section (below).

## Test

To run tests to test the functionality of command scripts:
```
go test
```

The test will create a tempt directory to run test commands in, and then clean up once done.

Currently, only the `snap` command is tested.

## Snap Command

To snap a webcam photo:
```
snapr snap
snapr snap --help
snapr snap --device /dev/video1
snapr snap --dir my/base/dir
snapr snap --extra-dir my/sub/dir
snapr snap --users
snapr snap --format png
```

## Upload Command

To upload a photo to an AWS bucket:
```
snapr upload --dir my/base/dir
snapr upload --file my/in/file.ext
```
