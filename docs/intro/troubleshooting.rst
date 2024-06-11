.. _troubleshooting:

Troubleshooting
===============

The first stop in any troubleshooting should always be to review logs. Whether
you understand everything you read or barely a word, it will give you at least
a few key words to start searching with.

Inspect ffmpeg
--------------

With ``loglevel`` set to ``DEBUG`` in ``config.yml``, the constructed ffmpeg
will be logged, prefixed with: ``Constructed ffmpeg command``:

  ::

    echo 'loglevel: DEBUG' >>config.yml
    python3 -m apr -a monitor
    
    DEBUG:Constructed ffmpeg command: ffmpeg -y -loglevel error -nostdin -nostats -f v4l2 -video_size 1920x1080 -framerate 5 -thread_queue_size 1024 -i /dev/video0 -f alsa -thread_queue_size 1024 -i hw:CARD=Generic_1,DEV=0 -vf drawtext=fontfile=/usr/share/fonts/truetype/freefont/FreeMonoBold.ttf:text="%{localtime}":fontcolor=red@0.8:x=7:y=7 -preset medium -t 00:01:30

Simply copy/paste, add ``filename.mkv`` to the end.

Python Debugger
---------------

When threading is not a concern, pudb is the absolute best python debugger.

Replace ``python3`` with ``pudb3`` to use:

  ::

    pudb3 -m apr -a monitor
