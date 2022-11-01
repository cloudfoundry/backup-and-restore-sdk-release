# boshcompiler

The `boshcompiler` script exists to emulate the Bosh compiler when creating "dockerized" Bosh releases.
These are container images used as a Bosh VM fake for faster testing.

The `boshcompiler` performs the following actions:
- It reads the `mainfest.MF` file that describes the Bosh release, reading the packages that need to be compiled and their dependencies.
- It compiles the packages in parallel, ensuring that any dependencies of a package have finished compiling first.

By compiling in parallel, the compilation time is 30-40% less than it would be in serial for the BBR SDK.
The efficiency gain will vary by Bosh release.

Usage:
```
  boshcompiler <Bosh release directory>
```

The `boshcompiler` expects to find a `Manifest.MF` and a `packages/` directory at the top level.
This corresponds with an extracted Bosh release tarball.