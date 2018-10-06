# go-sync presentation

go-sync or gsync is a CLI tool written in go to mirror 2 directories on the filesytem

Why such a tool when there's rsync or other ?

Well, using rsync in Cygwin on Windows caused me many issues (unreadable files) due to permissions and ownership.
Especially with NTFS filesystem when files are copied from 1 machine and accessed from another Windows system
(to be honest I did not really investigate the mitigations)

If you are on Linux/Unix, you'd better use rsync

# Limitations

* does not follow links, neither create them

# Build

* Install go>=1.11
* Run the Makefile `make`
* The binary is copied into the bin directory

# Usage

Example, mirror folder source1 and source2 into the destination directory destDir in dry-run mode
```
gsync -s source1 -s source2 -m -d destDir --dry-run
```

Then
```
gsync -s source1 -s source2 -m -d destDir
```
