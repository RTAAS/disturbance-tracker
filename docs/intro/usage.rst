.. _usage:

How to Use APR
==============

View help text:

  ::

    python3 -m apr --help

Begin continuous recording with:

  ::

    python3 -m apr -a monitor

Stop recording with:

  ::

    # From the same terminal session (will likely corrupt last video)
    Ctrl+C

    # Signal to finish recording and exit
    python3 -m apr -a monitor -s

    # Signal to finish recording and wait for process to exit
    python3 -m apr -a monitor -S

    # Signal to stop immediately (will likely corrupt last video)
    python3 -m apr -a monitor -H
