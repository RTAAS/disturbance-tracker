#!/usr/bin/make -f
##
# Makefile for Debian Package Checker (Review System)
##

monitor:
	python3 -m apr -a monitor

docs:
	make -C docs html
	# sensible-browser docs/_build/html/index.html

test:
	python3 -m pytest

clean:
	# Sphinx-doc
	make -C docs clean
	# Python cache
	find . -name '*.pyc' \
		-o -type d -name '__pycache__' \
		-o -type d -name '.pytest_cache' \
		-exec rm -rf {} \;

.PHONY: monitor app docs test clean
