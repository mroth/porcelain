# porcelain 🫖
[![Go Reference](https://pkg.go.dev/badge/github.com/mroth/porcelain/statusv2.svg)](https://pkg.go.dev/github.com/mroth/porcelain/statusv2)
<!-- [![CodeFactor](https://www.codefactor.io/repository/github/mroth/porcelain/badge)](https://www.codefactor.io/repository/github/mroth/porcelain) -->
<!-- [![Build Status](https://github.com/mroth/porcelain/workflows/test/badge.svg)](https://github.com/mroth/porcelain/actions) -->
<!-- [![codecov](https://codecov.io/gh/mroth/porcelain/branch/main/graph/badge.svg)](https://codecov.io/gh/mroth/porcelain) -->

Porcelain provides parsers for working with Git's [porcelain status output] in Go.

  - `porcelain=v1` is currently  not implemented.
  - [github.com/mroth/porcelain/statusv2] implements `porcelain=v2` formats.


It is unclear whether support for `porcelain=v1` will be added in the future.
Rudimentary support for v1 "-z format" is currently used in the internals for
[github.com/mroth/scmpuff], but it is a historic format with [lots of warts].

Support for the `porcelain=v2` format was first introduced in Git v2.11.0 (2016).

[porcelain status output]: https://git-scm.com/docs/git-status#_porcelain_format_version_2
[github.com/mroth/porcelain/statusv2]: https://pkg.go.dev/github.com/mroth/porcelain/statusv2
[github.com/mroth/scmpuff]: https://github.com/mroth/scmpuff
[lots of warts]: https://public-inbox.org/git/20100409184608.C7C61475FEF@snark.thyrsus.com/
