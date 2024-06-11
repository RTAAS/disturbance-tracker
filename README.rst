Audio Pattern Ranger
====================

Audio Pattern Ranger (APR) offers 24/7 monitoring for local disturbances
in an environment, using machine learning models to detect and log specific
nuisances, such as barking or car alarms. These models are trained on
collected data to automate logging of detected disturbances.

**Documentation**: https://audio-pattern-ranger.github.io/apr/

**Quickstart**::

    git clone https://github.com/audio-pattern-ranger/apr
    cd apr
    cp example_config.yml config.yml
    sensible-editor config.yml
    python3 -m apr --help

Background
----------

In some jurisdictions, understaffing can lead to a lack of support for
situations that are not life-threatening. In these cases, noise disturbances
may be entirely ignored without an extended log of repeated violation along
with video evidence proving log accuracy.

The primary purpose of this application is to simplify the collection and
analysis of video footage to identify disturbances (e.g., dog barks) using
a locally trained model. This model is designed to accurately detect and
classify specific disturbances in the local area.

How It Works
------------

1. Use the Monitor to collect some sample recordings
2. Dissect these recordings and extract individual noises (i.e. barks)
3. Use this data to (re-)train a machine learning model
4. Verify detection using original source clip
5. Use the Monitor to maintain continuous loop of recordings
6. Monitoring will scan completed recordings for trained noises
7. Use provided 'at' templates to auto-retain source data

Dependencies
------------

Debian::

    # Required
    apt-get install python3-fasteners ffmpeg v4l-utils

    # Recommended
    apt-get install fonts-freefont-ttf
