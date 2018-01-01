# go-sync presentation

go-sync or gsync is a CLI tool written in go to mirror 2 directories on the filesytem

Why such a tool when there's rsync or other ?

Well, using rsync in Cygwin on Windows caused me many issues (unreadable files) due to permissions and ownership.
Especially with NTFS filesystem when files are copied from 1 machine and accessed from another Windows system
(to be honest I did not really investigate the mitigations)

If you are on Linux/Unix, you'd better use rsync

# Limitations

* does not follow links, neither create them

# Usage

TBD