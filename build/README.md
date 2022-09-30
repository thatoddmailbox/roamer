## Building bottle
_first do the regular build process and make a release on GitHub_

Homebrew's built-in stuff to create bottles is pretty terrible, so it's replaced with this shell script. (which is also terrible in its own different way)

First. you need to update the Homebrew tap's formula to correspond to the new version and SHA256 hash. Then, copy that formula into this folder and delete its "bottle do" section. After that, you can run `./build-bottle.sh $version`.

Once that's done, you will have bottles for different OS versions in `bottleout`. The SHA256 hashes (which will all be the same) will be written to stdout, and you can copy those into the formula's `bottle do` section.