.. _installation:

Installation
============

APR can be installed on any unix-like system supporting Python 3.

.. _install-binary:

Binary
------

There is not currently enough interest/demand to generate installation packages.

.. _install-source:

Source
------

Follow this section to run APR from a source code checkout.

**Dependencies**

Debian:

  ::

    # Required
    apt-get install python3-fasteners ffmpeg v4l-utils

    # Recommended
    apt-get install fonts-freefont-ttf


**Repository**

Clone git repository:

  ::

    git clone https://github.com/audio-pattern-ranger/apr

If APR is "installed" using ``git clone``, then all commands must be executed
from within this directory. Navigate to it using ``cd apr``.

.. _install-verification:

Verification
------------

Successful installation can be verified by viewing help text:

  ::

    python3 -m apr --help
