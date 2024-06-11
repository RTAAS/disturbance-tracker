#!/usr/bin/env python3
# -*- coding: utf-8 -*-
'''
APR Tests
'''
import pytest
import warnings


def pytest_addoption(parser):
    '''
    Add extra options to pytest command. See: ``pytest --help``
    '''
    parser.addoption(
            '--skip-slow', action='store_true', default=False,
            help='Skip slow/time-intensive tests, such as test.sleep().')


def pytest_collection_modifyitems(config, items):
    '''
    Modify collected test stack.
    '''
    # Add '@pytest.mark.slow' decorator to test function to invoke skip.
    skip_slow = pytest.mark.skip(reason='Skip: Long Running')
    # Add '@pytest.mark.broken' decorator to test function to invoke skip.
    skip_broken = pytest.mark.skip(
            reason='Skip: Broken | ONLY FOR DEVELOPING TESTS')

    for item in items:
        if 'broken' in item.keywords:
            warnings.warn(UserWarning('WARNING: BROKEN TEST(S) SKIPPED'))
            item.add_marker(skip_broken)
        if 'slow' in item.keywords and config.getoption('--skip-slow'):
            item.add_marker(skip_slow)
