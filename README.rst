Disturbance Tracker
===================

``DTrack`` is Surveilance Software that uses Machine Learning (AI/ML) to
learn about and report on local disturbances, without nvidia or aws.

**Documentation:** https://mtecknology.github.io/dtrack/

Background
----------

Various excuses were being used to ignore reports of nuisance barking. The only
solution was to wait for at least one hour of nuisance barking across two days
and then build a detailed log that tracked every minute that a dog barked.

Similar solutions to this problem existed, but generic solutions fell horribly
short (confusing wind with dog barks) and targeted solutions required high-cost
resources (AWS + always-on internet) to operate.

``DTrack`` attempts to solve the same problem without external resources.

How It Works
------------

  .. image:: https://raw.githubusercontent.com/MTecknology/dtrack/refs/heads/master/.github/images/workflow.webp
     :alt: Disturbance Tracker Workflow

1. Set up ``Monitoring Device``
2. Collect some initial recordings
3. Manually review recordings and tag nuisance noises
4. Train a model (i.e. "Teach AI")
5. Use trained model for automatic detection

Monitoring Device
~~~~~~~~~~~~~~~~~

DTrack is designed to run on a ``Raspberry Pi v5``. It does not require any
internet connection or subscription-style service.


Configuration Options
---------------------

Configuration is stored in a ``JSON`` file, called ``config.json``. A sample can
be copied from ``example_config.json``.

Full configuration options are available using ``./dtrack --help``.

**TODO:** ``docs -> github -> detailed config option docs``
