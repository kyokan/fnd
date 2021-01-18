#!/usr/bin/env bash

set -e

root_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."
build_dir="$root_dir/build"
tag="$(git describe --tags --abbrev=0)"
version="${tag:1}"

create_tarball() {
  mv -f "$build_dir/fnd-$1-amd64" "$build_dir/fnd"
  mv -f "$build_dir/fnd-cli-$1-amd64" "$build_dir/fnd-cli"
  tar -czvf "$build_dir/fnd-$version-$1-amd64.tgz" -C "$build_dir" "fnd" "fnd-cli"
  gpg2 --detach-sig --default-key D4B604F1 --output "$build_dir/fnd-$version-$1-amd64.tgz.sig" "$build_dir/fnd-$version-$1-amd64.tgz"
}

upload_binary() {
  echo "Uploading $1 binary..."
  gothub upload --user fnd-org --repo fnd --tag "$tag" --file "$build_dir/fnd-$version-$1-amd64.tgz" --name "fnd-$version-$1-amd64.tgz"
  gothub upload --user fnd-org --repo fnd --tag "$tag" --file "$build_dir/fnd-$version-$1-amd64.tgz.sig" --name "fnd-$version-$1-amd64.tgz.sig"
}

make clean
# package-deb runs make all internally
make package-deb version="$version"

create_tarball "linux"
create_tarball "darwin"
gpg2 --detach-sig --default-key D4B604F1 --output "$build_dir/fnd-$version-amd64.deb.sig" "$build_dir/fnd-$version-amd64.deb"

gothub release --user fnd-org --repo fnd --tag "$tag" --name "$tag" --description ""
upload_binary "linux"
upload_binary "darwin"
echo "Uploading .deb..."
gothub upload --user fnd-org --repo fnd --tag "$tag" --file "$build_dir/fnd-$version-amd64.deb" --name "fnd-$version-amd64.deb"
gothub upload --user fnd-org --repo fnd --tag "$tag" --file "$build_dir/fnd-$version-amd64.deb.sig" --name "fnd-$version-amd64.deb.sig"

# for seed node deployments

s3cmd put "$build_dir/fnd-$version-linux-amd64.tgz" s3://fnd-releases/fnd-linux-amd64.tgz
s3cmd setacl s3://fnd-releases/fnd-linux-amd64.tgz --acl-public
s3cmd put "$build_dir/fnd-$version-linux-amd64.tgz.sig" s3://fnd-releases/fnd-linux-amd64.tgz.sig
s3cmd setacl s3://fnd-releases/fnd-linux-amd64.tgz.sig --acl-public
cp "$build_dir/fnd-$version-linux-amd64.tgz" "$build_dir/fnd-linux-amd64.tgz"
cd "$build_dir" && shasum -a 256 "fnd-linux-amd64.tgz" > /tmp/fnd-linux-amd64.tgz.sum.txt && cd "$DIR"
s3cmd put /tmp/fnd-linux-amd64.tgz.sum.txt s3://fnd-releases/fnd-linux-amd64.tgz.sum.txt
s3cmd setacl s3://fnd-releases/fnd-linux-amd64.tgz.sum.txt --acl-public
