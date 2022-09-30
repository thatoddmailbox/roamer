#!/bin/bash
set -eo pipefail

if [[ -z $1 ]]; then
	echo "You need to specify a version."
	exit 1
fi

version=$1
os_versions=(
	big_sur
	arm64_big_sur
	monterey
	arm64_monterey
	ventura
	arm64_ventura
)

! rm -r bottlewd 2>/dev/null
mkdir bottlewd
! rm -r bottleout 2>/dev/null
mkdir bottleout

bottlecontents=bottlewd/roamer/$version
mkdir -p $bottlecontents

cp ../README.md $bottlecontents
cp ../LICENSE $bottlecontents

source_modified_time=`git show -s --format=%ct`
echo -n '{"homebrew_version":"2.3.0","used_options":[],"unused_options":[],"built_as_bottle":true,"poured_from_bottle":false,"installed_as_dependency":false,"installed_on_request":true,"changed_files":["INSTALL_RECEIPT.json"],"time":null,"source_modified_time":'$source_modified_time',"HEAD":null,"stdlib":null,"compiler":"clang","aliases":[],"runtime_dependencies":[],"source":{"path":"@@HOMEBREW_REPOSITORY@@/Library/Taps/thatoddmailbox/homebrew-tap/Formula/roamer.rb","tap":"thatoddmailbox/tap","spec":"stable","versions":{"stable":"'$version'","devel":"","head":"","version_scheme":0}}}' > $bottlecontents/INSTALL_RECEIPT.json

mkdir -p $bottlecontents/bin
builddir=`pwd`
pushd $bottlecontents/bin
tar -xf $builddir/output/roamer_${version}_darwin_universal.tar.gz
popd

mkdir -p $bottlecontents/.brew
cp roamer.rb $bottlecontents/.brew

tar -f bottle.tar.gz -c -z -C bottlewd roamer

for i in "${os_versions[@]}"
do
	cp bottle.tar.gz bottleout/roamer-${version}.$i.bottle.tar.gz
done

shasum -a 256 bottleout/*