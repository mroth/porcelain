# porcelain 🫖
[![Go Reference](https://pkg.go.dev/badge/github.com/mroth/porcelain/statusv2.svg)](https://pkg.go.dev/github.com/mroth/porcelain/statusv2)
<!-- [![CodeFactor](https://www.codefactor.io/repository/github/mroth/porcelain/badge)](https://www.codefactor.io/repository/github/mroth/porcelain) -->
<!-- [![Build Status](https://github.com/mroth/porcelain/workflows/test/badge.svg)](https://github.com/mroth/porcelain/actions) -->
<!-- [![codecov](https://codecov.io/gh/mroth/porcelain/branch/main/graph/badge.svg)](https://codecov.io/gh/mroth/porcelain) -->

Porcelain provides parsers for Git's [porcelain status output] in Go.

  - `porcelain=v1` is currently  not implemented.
  - [github.com/mroth/porcelain/statusv2] implements `porcelain=v2` formats.

The parsers are quite performant (parsing a typical git status report including
headers in ~2µs single-threaded), and fuzz tested to avoid any possible crashing
panics.

The `porcelain=v2` format was first introduced in Git v2.11.0 (2016).

It is undetermined whether legacy support for `porcelain=v1` will be added to
this library in the future. Rudimentary support for v1 "-z format" is currently
used in the internals for my [github.com/mroth/scmpuff] project, but it is a
historic format with [some issues].

[porcelain status output]: https://git-scm.com/docs/git-status#_porcelain_format_version_2
[github.com/mroth/porcelain/statusv2]: https://pkg.go.dev/github.com/mroth/porcelain/statusv2
[github.com/mroth/scmpuff]: https://github.com/mroth/scmpuff
[some issues]: https://public-inbox.org/git/20100409184608.C7C61475FEF@snark.thyrsus.com/
