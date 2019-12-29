
# Snapr CLI

Snaps a webcam image using `ffmpeg`.

Includes a command that enables you to securely browse a Private or Public S3 Bucket in your browser. 

Everything works on Linux and Mac computers. 
Everything except `upload` and `webcam` works on Windows. 

## Environment

To successfully run this code, you need a file at the project root named `.env`, and yes, it can be blank.
After building, you will no longer need the `.env` file.
For the environment variables specified in the `.env` file, after building, they will no longer be overridable at runtime, except by use of another `.env` file. 
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

## Global Flags

The following may be overridden via command line ofr `any command`:
```
--s3-bucket=abc
--s3-region=bca
--s3-token=xyz
--s3-secret=yzx
```

## Snap Command

To `snap` a webcam or screenshot photo:
```
snapr snap
snapr snap --help
snapr snap --dir=my/base/dir
snapr snap --dir=my/base/dir --device=/dev/video1
snapr snap --dir=my/base/dir --extra-dir=my/sub/dir
snapr snap --dir=my/base/dir --users
```

To trigger the photo(s) & upload to an S3 bucket at the same time:
```
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


## Download Command

To `download` from a bucket to your computer:
```
snapr download --s3-key=path/to/origs --s3-is-dir
snapr download --s3-key=path/to/origs/original.ext --work-dir /home/my/desktop
```

These OPTIONAL `.env` vars apply to the `rename` command flags:
```
DOWNLOAD_WORK_DIR=
DOWNLOAD_S3_KEY=
DOWNLOAD_S3_IS_DIR=
```

## Upload Command

To `upload` a file or directory to an AWS bucket:
```
snapr upload --file=my/in/file.ext 
snapr upload --file=my/in/file.ext --s3-dir dir/in/s3
snapr upload --file=my/in/file.ext --cleanup
snapr upload --dir=my/base/dir 
snapr upload --dir=my/base/dir --s3-is-public
snapr upload --dir=my/base/dir --limit=10
snapr upload --dir=my/base/dir --limit=10 --formats=png,jpg
```

If `--formats` is not specified, then all files are uploaded.

If `--s3-is-public` is not specified, then all files are `private`.

These OPTIONAL `.env` vars apply to the `upload` command flags:
```
UPLOAD_DIR=
UPLOAD_FILE=
UPLOAD_CLEANUP_AFTER_SUCCESS=
UPLOAD_FORMATS=
UPLOAD_LIMIT=
UPLOAD_S3_DIR=
UPLOAD_S3_IS_PUBLIC=
```

## Delete Command

To `delete` a file or directory from an AWS bucket:
```
snapr delete --s3-key=path/to/object.ext 
snapr delete --s3-key=path/to/dir --is3-s-dir
```

These OPTIONAL `.env` vars apply to the `delete` command flags:
```
DELETE_S3_KEY=
DELETE_S3_IS_DIR=
```

## Rename / Copy Command

To `rename (or copy)` from one bucket to another (or same bucket):
```
snapr rename --s3-src-key=path/to/original.ext --s3-dest-key=path/to/dest.ext
snapr rename --s3-src-key=path/to/original.ext --s3-dest-key=path/to/dest.ext --s3-dest-bucket other-bucket --copy
snapr rename --s3-src-key=path/to/orig --src-is-dir --s3-dest-key=path/to/dest
snapr rename --s3-src-key=path/to/orig --src-is-dir --s3-dest-key=path/to/dest --s3-dest-bucket other-bucket --copy
snapr rename --copy --s3-src-key=originals --s3-dest-bucket=my.public.bucket --s3-dest-key=originals --s3-dest-is-public --s3-src-is-dir
```

These OPTIONAL `.env` vars apply to the `rename` command flags:
```
RENAME_S3_SRC_KEY=
RENAME_S3_DEST_KEY=
RENAME_S3_DEST_BUCKET=
RENAME_SRC_IS_DIR=
RENAME_IS_COPY_OPERATION=
```

# TODO: Doc process command

Rebuild all assets:
```
snapr process --s3-dest-key=processed --s3-is-public --s3-src-key=originals --sizes=640,728,1024
```

Rebuild new assets (based on scanning of the output dirs):
```
snapr process --s3-dest-key=processed --s3-is-public --s3-src-key=originals --sizes=640,728,1024
snapr process --s3-dest-key=processed --s3-is-public --s3-src-key=originals --sizes=640,728,1024 --rebuild-new 
```

## Serve Command

WORK IN PROGRESS

Used to view files from a public or private S3 Bucket in the browser on your local computer, and also to manage files in multiple buckets.

Use it like this:
```
snapr serve --port=8081
snapr serve --work-dir=/Users/me/Desktop
```

Then, in a browser, go to the address the CLI indicates.
BE VERY CAREFUL what you click :)
With great power comes great responsibility!
Adding more functions soon ... 

These OPTIONAL `.env` vars apply to the `serve` command flags:
```
SERVE_WORK_DIR=
SERVE_PORT=
```

## TODO

- LOTS of TODOs in the code
- serve - add rotate function
- Make the tests clean up themselves on the s3?
- Todo Permissions override for mkdir functionality
- serve command - view file as text (for text file types)
- serve command - add upload capability from ui
- Prod build - no debugging statements with sensitive info, DO NOT allow override of env provided from binary package
- Dev build - allow override of env provided from binary package
- Limited Access build - Prod build, plus no serve command?
- Test and document with PAM and Crontab (exit code 0 for pam)
- Add Device List Command and tests to list capture devices
- Make Webcam and upload work on windows
- serve command - add soft delete capability (batch?)


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
