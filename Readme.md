
# Snapr CLI

Snaps a webcam image using `ffmpeg`.
Works on Linux and Mac computers.

## TODO

- Add Serve command functionality
- Test and document with PAM and Crontab (exit code 0 for pam)
- Todo Permissions override for mkdir functionality
- AWS Upload - do not clean up if not successful
- Add List Command and tests to list capture devices

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
go test -run="SnapC"
```

To run just the `upload` command tests:
```
go test -run="UploadC"
```

To test a combination of the `snap` and `upload` commands:
```
go test -run="SnapU"
```

The `serve` command is not currently tested.

## Snap Command

To snap a webcam photo:
```
snapr snap
snapr snap --help
snapr snap --device=/dev/video1
snapr snap --dir=my/base/dir
snapr snap --extra-dir=my/sub/dir
snapr snap --users
snapr snap --format=png
snapr snap --upload --cleanup
```

## Upload Command

To upload a photo to an AWS bucket:
```
snapr upload --file=my/in/file.ext
snapr upload --file=my/in/file.ext --cleanup
snapr upload --dir=my/base/dir 
snapr upload --limit=10
```

## Serve Command

WORK IN PROGRESS

Used to view files from a public or private S3 Bucket in the browser on your local computer.

Use it like this:
```
snapr serve --s3-dir=cupcake
snapr serve --port=8081
```
