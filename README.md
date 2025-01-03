# condapaths
condapaths parses wrstat stats.gz files quickly, in low mem.

Provide a directory as an argument and it will parse the most recent stats.gz
file inside.

It outputs files with one quoted path per line:
* <date>.condarc: paths where the file basename was ".condarc"
* <date>.conda-meta: paths where the file basename was "history", in a directory
                     named "conda-meta"
* <date>.singularity: paths where the file basename suffix was one of ".sif",
                      ".simg", and ".img"

Usage: condapaths /wrstat/output/dir
