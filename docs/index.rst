.. _index:

.. toctree::
   :hidden:
   :includehidden:

   About APR <self>
   Installation <intro/install>
   Configuration <intro/configure>
   How to Use <intro/usage>
   Troubleshooting <intro/troubleshooting>

.. _apr:

About APR
=========

**Audio Pattern Ranger (APR)** offers 24/7 monitoring for local disturbances
in an environment, using machine learning models to detect and log specific
nuisances, such as barking or car alarms. These models are trained on
collected data to automate logging of detected disturbances.

.. _why:

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

.. _how-it-works:

How It Works
------------

  1. Use the Monitor to collect some sample recordings
  2. Dissect these recordings and extract individual noises (i.e. barks)
  3. Use this data to (re-)train a machine learning model
  4. Verify detection using original source clip
  5. Use the Monitor to maintain continuous loop of recordings
  6. Monitoring will scan completed recordings for trained noises
  7. Use provided 'at' templates to auto-retain source data

.. _getting-started:

Getting Started
---------------

  1. :ref:`Install APR <installation>`
  2. :ref:`Configure <configuration>`
  3. :ref:`Basic Usage <usage>`
