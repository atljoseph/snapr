
# This script builds this project in all supported operating systems

echo 'Build in all platforms'

echo 'Cleanup Dir: bin/*'

# remove the current bin output
rm -rf bin/*

# ensure dir
mkdir -p bin/prod

echo 'Initializing pkger files'

# clean up the pkger output

rm -rf pkged.go

echo 'Cleaned pkger'

# pkger - include files from 
pkger -include snapr:/.env

echo 'Included files w/ pkger'

# build each output and move into the appropriate folder

# linux
# ==============================================
echo 'Build linux'

# build
env GOOS=linux go build

# ensure dir
mkdir -p bin/prod/linux

# move binary
mv snapr bin/prod/linux/

# darwin
# ==============================================
echo 'Build darwin'

# build
env GOOS=darwin go build

# ensure dir
mkdir -p bin/prod/darwin

# move binary
mv snapr bin/prod/darwin/

# windows
# ==============================================
echo 'Build windows'

# build
env GOOS=windows go build

# ensure dir
mkdir -p bin/prod/windows

# move binary
mv snapr.exe bin/prod/windows/

echo 'Done building'

# clean up the pkger output

rm -rf pkged.go

echo 'Done cleaning pkger files'
