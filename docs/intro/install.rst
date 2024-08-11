.. _installation:

Installation
============

APR can be installed on any unix-like system supporting Python 3.

.. _install-binary:

Binary
------

To Do?

.. _install-source:

Source
------

Follow this section to run APR from a source code checkout.

**Dependencies**

Debian:

.. code-block:: sh

    # Required
    apt-get install ffmpeg v4l-utils python3-virtualenv
    #python3-fasteners

    # Recommended
    apt-get install fonts-freefont-ttf

**Py3 VirtualEnv**

TODO: Is this needed??

Create an initial environment with:

.. code-block:: sh

    python3 -m venv ~/.mlpy

and "activate" with:

.. code-block:: sh

    .  ~/.mlpy/bin/activate

Install python dependencies:

.. code-block:: sh

    pip3 install -r requirements.txt

**Source Code**

Clone git repository:

.. code-block:: sh

    git clone https://github.com/audio-pattern-ranger/apr

or choose `Download ZIP <https://github.com/audio-pattern-ranger/apr>`__ and unzip.

If APR is "installed" using ``git clone``, then all commands must be executed
from within ``./apr``.

.. _install-verification:

Verification
------------

Successful installation can be verified by viewing help text:

.. code-block:: sh

    (.mlpy) michael@vsense1:~/apr $ python3 -m apr --help
    usage: apr [-h] -a <action> [other_options]
    [...]
