# porcelain 🫖
[![Go Reference](https://pkg.go.dev/badge/github.com/mroth/porcelain/statusv2.svg)](https://pkg.go.dev/github.com/mroth/porcelain/statusv2)
<!-- [![CodeFactor](https://www.codefactor.io/repository/github/mroth/porcelain/badge)](https://www.codefactor.io/repository/github/mroth/porcelain) -->
<!-- [![Build Status](https://github.com/mroth/porcelain/workflows/test/badge.svg)](https://github.com/mroth/porcelain/actions) -->
<!-- [![codecov](https://codecov.io/gh/mroth/porcelain/branch/main/graph/badge.svg)](https://codecov.io/gh/mroth/porcelain) -->

Porcelain provides parsers for Git's [porcelain status output] in Go.

  - [github.com/mroth/porcelain/statusv1] provides `porcelain=v1` format parsing.
  - [github.com/mroth/porcelain/statusv2] provides `porcelain=v2` format parsing.

The parsers are performant (parsing a typical git status report including
headers in ~2µs single-threaded), and robust (fuzz tested to avoid any possible
crashing panics).

Support for both regular (LF delimited) and `-z` (NUL delimited) output formats
is provided.

The `porcelain=v2` format was first introduced in Git v2.11.0 (2016), and is
recommended for most use cases, as it is significantly more robust and addresses
[some inconsistencies] with the historic `porcelain=v1` format.

[porcelain status output]: https://git-scm.com/docs/git-status#_porcelain_format_version_2
[github.com/mroth/porcelain/statusv1]: https://pkg.go.dev/github.com/mroth/porcelain/statusv1
[github.com/mroth/porcelain/statusv2]: https://pkg.go.dev/github.com/mroth/porcelain/statusv2
[github.com/mroth/scmpuff]: https://github.com/mroth/scmpuff
[some inconsistencies]: https://public-inbox.org/git/20100409184608.C7C61475FEF@snark.thyrsus.com/
