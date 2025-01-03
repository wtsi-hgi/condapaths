# condapaths
condapaths parses wrstat stats.gz files quickly, in low mem.

Provide one or more stats.gz files output by wrstat.

It outputs files with one path per line:
* <input prefix>.condarc: paths where the file basename was ".condarc"
* <input prefix>.conda-meta: paths where the file basename was "history", in a
                             directory named "conda-meta"
* <input prefix>.singularity: paths where the file basename suffix was one of
                              ".sif",  ".simg", and ".img"

Usage: condapaths 20241222_mount.unique.stats.gz
