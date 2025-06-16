.PHONY: docs/git-status.txt
docs/git-status.txt:
	git status --help | col -b > $@
