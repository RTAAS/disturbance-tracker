.. _configuration:

Configuration
=============

The sample configuration file shows all options and their default options.
Start with a copy of this file and then modify it as needed.

.. code-block:: sh

    cp example_config.yml config.yml
    sensible-editor config.yml

The only value that **must** be set is the audio capture device. If the
video capture device is set to ``None``, only audio will be recorded.

**Audio capture** devices and their capabilities can be listed using:

.. code-block:: sh

    ffmpeg -loglevel warning -sources alsa

Sample Output:

.. code-block:: sh

    Auto-detected sources for alsa:
    null [Discard all samples (playback) or generate zero samples (capture)]
    hw:CARD=Snowflake,DEV=0 [Direct hardware device without any conversions]
    plughw:CARD=Snowflake,DEV=0 [Hardware device with all software conversions]
    default:CARD=Snowflake [Default Audio Device]
    sysdefault:CARD=Snowflake [Default Audio Device]
    front:CARD=Snowflake,DEV=0 [Front output / input]
    dsnoop:CARD=Snowflake,DEV=0 [Direct sample snooping device]

In the above example, ``hw:CARD=Snowflake,DEV=0`` is one valid string that can
be used for configuration. It is wise to test all available options to determine
which will yield the most complete result.

**Video capture** devices and their capabilities can be listed using:

.. code-block:: sh

    v4l2-ctl --list-devices --all

Sample Output:

.. code-block:: sh

    Integrated Camera: Integrated C (usb-0000:64:00.4-1):
            /dev/video0
            /dev/video1
            /dev/video2
            /dev/media0
            /dev/media1

    Lenovo 500 RGB Camera: Lenovo 5 (usb-0000:01:00.0-1.2):
            /dev/video3
            /dev/video4
            /dev/media2

    [...]

    Format Video Capture:
            Width/Height      : 1920/1080
            Pixel Format      : 'MJPG' (Motion-JPEG)
            [...]
    Streaming Parameters Video Capture:
            Capabilities     : timeperframe
            Frames per second: 5.000 (5/1)

    [...]

This sample output provided shows a webcam that can record at a maximum
resolution of 1920x1080 at a framerate of 5. Note that only "video" devices
should be used as a video capture device.
