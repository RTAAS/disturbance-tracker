.. _usage:

How to Use APR
==============

APR is essentially broken into four phases:

1. Data Collection (monitor without model)
2. Data Analysis (inspection)
3. Model Training (train)
4. Auto Analysis (monitor with model)

View help text:

.. code-block:: sh

    python3 -m apr --help

Data Collection
---------------

The entire process starts with collecting some initial data. This data should be
collected from :ref:`the recording device <record>` using the same hardware and
recording settings that will be used for regular analysis.

Begin continuous recording with:

.. code-block:: sh

    python3 -m apr -a monitor

Stop recording with:

.. code-block:: sh

    # From the same terminal session (will likely corrupt current video)
    Ctrl+C

    # Signal to finish recording and exit
    python3 -m apr -a monitor -s

    # Signal to finish recording and wait for process to exit
    python3 -m apr -a monitor -S

    # Signal to stop immediately (will likely corrupt last video)
    python3 -m apr -a monitor -H

Recordings will be saved to ``./_workspace/rotating/``.

.. note::

   Clapping hands together is a great demonstration exercise. This can be set
   in ``config.yml`` with ``models: [clap]``.

Data Analysis
-------------

During initial data collection, it can be useful to set ``record_duration:`` to
2-5 minutes and then rename each recording as they complete, using the following
as an example:

.. code-block:: text

    2024-08-10_13:54:00_TRAIN-clap.mkv
    2024-08-10_13:58:17_TRAIN-clap.mkv
    2024-08-10_14:00:26_TEST-nomatch.mkv
    2024-08-10_14:02:38_TEST-nomatch.mkv
    2024-08-10_14:04:57_TEST-nomatch.mkv
    2024-08-10_14:07:20_TEST-nomatch.mkv
    2024-08-10_14:09:45_TEST-clap.mkv
    2024-08-10_14:17:07_TEST-nomatch.mkv
    2024-08-10_14:19:35_TEST-nomatch.mkv
    2024-08-10_14:22:02_TEST-clap.mkv

- ``TRAIN`` files were created with many variations of the sound being searched
  for, using different background noise, volumes, etc. These files will be cut
  into 1-second clips for model training.
- ``TEST`` files include variation, but may have only one instance of a sound
  being searched for--one needle in the haystack. These will be used to test the
  quality of each ML iteration.

Once data is collected, it can be retrieve from ``_workspace/rotating`` on the
:ref:`recording device <record>` and copied to the same ``_workspace/rotating``
location on the :ref:`device used for training <train>`.

**Testing Data:**

Test data for each model (or "nomatch") can be moved to
``_workspace/test/<model>/`` or ``_workspace/test/nomatch/``:

.. code-block:: sh

   mkdir -p _workspace/test/{nomatch,clap}
   mv _workspace/rotating/*_TEST-nomatch.mkv _workspace/test/nomatch/
   mv _workspace/rotating/*_TEST-clap.mkv _workspace/test/clap/

**Training Data:**

In order to determine if something ``is`` or ``is not``, the source audio must
be broken up into short consumable segments and segments matching the target
model must be reviewed and saved (tagged) manually.

.. admonition:: Project Timing

   - APR is designed for generating reports.
   - Report granularity uses 1-minute cycles.
     + 1 clap or 999 claps within 1 minute is logged as one hit.
   - Each recording is broken into 1-second clips.
   - Each clip overlaps the next by 0.1 seconds to prevent dead zones

Open and review captured (from ``rotating/``) using the inspection tool:

.. code-block:: sh

    python3 -m apr -a inspect

The ``inspect`` option provides a GUI to help simplify the process of reviewing
and tagging 1-second clips.

Keyboard Shortcuts:

- Left/Right: Navigate 1 frame left or right
- PgUp/PgDn: Navigate 60 frames left or right
- Home/End: Navigate to start or end
- Up: Replay audio clip
