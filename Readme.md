
# Snapr CLI

Snaps a webcam image using `ffmpeg`.

Includes a command that enables you to securely browse a Private or Public S3 Bucket in your browser. 

Works on Linux and Mac computers.

## TODO

- serve command - view file as text (for text file types)
- serve command - add move/rename capability (batch?)
- serve command - add soft/hard delete capability (batch?)
- serve command - add upload capability
- Make the tests clean up themselves on the s3?
- Prod build - no debugging statements with sensitive info, DO NOT allow override of env provided from binary package
- Dev build - allow override of env provided from binary package
- Limited Access build - Prod build, plus no serve command?
- cacpture screenshots with https://github.com/kbinani/screenshot
- Test and document with PAM and Crontab (exit code 0 for pam)
- Todo Permissions override for mkdir functionality
- Add Device List Command and tests to list capture devices

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
GOOS=windows go build snapr (`snap` command not supported ... yet)
```

To use a custom `.env` file:
```
go build -ldflags "-X main.EnvFilePath=.prod.env" snapr
```

## Environment

You need a file at the project root named `.env`, and yes, it can be blank.
For the environment variables specified in the `.env` file, after building, they will no longer be overridable at runtime. 
If a variable is not set, then it can be overriden at runtime.

These REQUIRED variables are used for AWS S3 access:
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

## User Permissions

Can run this as `sudo`, but also as other users.

If a user is not iin the `video` group, then they need to be added:
```
sudo adduser <user> video
```

Also, this can help with regards to storage direcory permissions when potentially multiple users:
```
sudo chmod -R 0777 /output/dir
```

Also helps tp have all your directories set up before hand with proper permissions.

## Testing

To run tests to test the functionality of command scripts:
```
go test
```

The test will create a temp directory to run test commands in, and then clean up once done.

To run just the `snap` command tests:
```
go test -run=1
```

To run just the `upload` command tests:
```
go test -run=2
```
The `--public` flag is not included in automated tested.

To test a combination of the `snap` and `upload` commands:
```
go test -run=3
```

The `serve` command is not currently tested.

## Snap Command

To snap a webcam photo:
```
snapr snap
snapr snap --help
snapr snap --dir=my/base/dir
snapr snap --dir=my/base/dir --device=/dev/video1
snapr snap --dir=my/base/dir --extra-dir=my/sub/dir
snapr snap --dir=my/base/dir --users
snapr snap --dir=my/base/dir --format=png --upload 
snapr snap --file test.jpg --upload --cleanup
snapr snap --file test.jpg --screen --upload --cleanup
snapr snap --file test.jpg --camera --screen --upload --cleanup
```

Please feel free to combine options.

These OPTIONAL `.env` vars apply to the `snap` command flags:
```
SNAP_SCREENSHOT=
SNAP_CAMERA=
SNAP_DEVICE=
SNAP_DIR_EXTRA=
SNAP_DIR=
SNAP_FILE=
SNAP_FILE_FORMAT=
SNAP_FILE_USERS=
SNAP_UPLOAD_AFTER_SUCCESS=
SNAP_CLEANUP_AFTER_UPLOAD=
```

## Upload Command

To upload a photo to an AWS bucket:
```
snapr upload --file=my/in/file.ext 
snapr upload --file=my/in/file.ext --s3-dir dir/in/s3
snapr upload --file=my/in/file.ext --cleanup
snapr upload --dir=my/base/dir 
snapr upload --dir=my/base/dir --public
snapr upload --dir=my/base/dir --limit=10
snapr upload --dir=my/base/dir --limit=10 --formats=png,jpg
```

If `--formats` is not specified, then all files are uploaded.

If `--public` is not specified, then all files are `private`.

These OPTIONAL `.env` vars apply to the `upload` command flags:
```
UPLOAD_DIR=
UPLOAD_FILE=
UPLOAD_CLEANUP_AFTER_SUCCESS=
UPLOAD_FORMATS=
UPLOAD_LIMIT=
UPLOAD_S3_DIR=
UPLOAD_PUBLIC=
```

## Serve Command

WORK IN PROGRESS

Used to view files from a public or private S3 Bucket in the browser on your local computer.

Use it like this:
```
snapr serve --s3-dir=cupcake
snapr serve --work-dir=/Users/me
snapr serve --port=8081
snapr serve --work-dir=/Users/me/Desktop --s3-dir=test
```

These OPTIONAL `.env` vars apply to the `serve` command flags:
```
SERVE_S3_DIR=
SERVE_WORK_DIR=
SERVE_PORT=
```