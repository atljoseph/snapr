
# Snapr CLI

I built this tool to help manage and process photo albums for my website, and for other things ...

Extras:
- Includes a command that enables you to securely browse a Private or Public S3 Bucket in your browser. 
- Capable of snapping a webcam image using `ffmpeg`, and / or a screenshot.

Everything works on Linux and Mac computers. 
Everything except `upload` and `webcam` works on Windows. 

## Environment

To successfully run this code, you need a file at the project root named `.env`, and yes, it can be blank.
After building, you will no longer need the `.env` file.
For the environment variables specified in the `.env` file, after building, they will no longer be overridable at runtime, except by use of another `.env` file or by explicitly available command line flag(s). 
If a variable is not set, then it can be overriden at runtime.

See `default.env` for an example env config file, and the required env variable names.

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

You will need the `pkger` cli to embed files in the binary:
```
go get github.com/markbates/pkger/cmd/pkger
pkger -h
```

Run this file to build all binaries in the bin folder:
```
sh build-all-platforms.sh 
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

The following may be overridden via command line for `any command`:
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

Review the code to discover environment variables related to this command.

## Download Command

To `download` from a bucket to your computer:
```
snapr download --s3-key=path/to/origs --s3-is-dir
snapr download --s3-key=path/to/origs/original.ext --work-dir /home/my/desktop
```

Review the code to discover environment variables related to this command.

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

Review the code to discover environment variables related to this command.

## Delete Command

To `delete` a file or directory from an AWS bucket:
```
snapr delete --s3-key=path/to/object.ext 
snapr delete --s3-key=path/to/dir --is3-s-dir
```

Review the code to discover environment variables related to this command.

## Rename / Copy Command

To `rename (or copy)` from one bucket to another (or same bucket):
```
snapr rename --s3-src-key=path/to/original.ext --s3-dest-key=path/to/dest.ext
snapr rename --s3-src-key=path/to/original.ext --s3-dest-key=path/to/dest.ext --s3-dest-bucket other-bucket --copy
snapr rename --s3-src-key=path/to/orig --src-is-dir --s3-dest-key=path/to/dest
snapr rename --s3-src-key=path/to/orig --src-is-dir --s3-dest-key=path/to/dest --s3-dest-bucket other-bucket --copy
snapr rename --copy --s3-src-key=originals --s3-dest-bucket=my.public.bucket --s3-dest-key=originals --s3-dest-is-public --s3-src-is-dir
```

Review the code to discover environment variables related to this command.

## Process command

Rebuild all assets:
```
snapr process --s3-dest-key=processed --s3-is-public --s3-src-key=originals --sizes=640,728,1024
```

Rebuild new assets (based on scanning of the output dirs):
```
snapr process --s3-dest-key=processed --s3-is-public --s3-src-key=originals --sizes=640,728,1024
snapr process --s3-dest-key=processed --s3-is-public --s3-src-key=originals --sizes=640,728,1024 --rebuild-new 
```

Review the code to discover environment variables related to this command.

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

Review the code to discover environment variables related to this command.

# User Permissions

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
