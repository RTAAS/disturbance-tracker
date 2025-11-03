'''
Disturbance Tracker - Bootstrap and Options
'''
import argparse
import json
import logging


# NOTE: Partially replicates "Application_Configuration" from src/config.go
DTRACK_DEFAULTS = {
        "train_target": 0.95,
        "train_rate": 0.001,
        "train_momentum": 0.9,
        "train_dropout": 0.2,
        }

# Currently loaded configuration options
_loaded_configuration = {}


def get(key):
    '''
    Return the value for a named key
    '''
    if not _loaded_configuration:
        bootstrap()
    return _loaded_configuration[key]


def bootstrap():
    '''
    Bootstrap process that loads command line arguments and configuration options
    '''
    opts = read_arguments()
    load_configuration(opts.config_path)
    if opts.very_verbose:
        level = 'TRACE'
    elif opts.verbose:
        level = 'DEBUG'
    else:
        level = 'INFO'
    configure_logging(level)


def load_configuration(path):
    '''
    Load configuration file and merge with defaults
    '''
    global _loaded_configuration
    with open(path) as fh:
        config = json.load(fh)
    _loaded_configuration = {**config, **DTRACK_DEFAULTS}


def configure_logging(log_level):
    '''
    Configure log format and output log level
    '''
    log = logging.getLogger()
    log.setLevel(log_level)
    log_format = logging.Formatter('%(levelname)s:%(message)s')
    for h in log.handlers:
        h.setFormatter(log_format)


def read_arguments():
    '''
    Returns all options read from argument parser
    '''
    parser = argparse.ArgumentParser(
            usage='dtrack [-h] ai.<action> [other_options]',
            formatter_class=lambda prog: argparse.HelpFormatter(
                prog, max_help_position=30))

    # NOTE: Replicates "Application Flags" from src/flags.go
    parser.add_argument(
        '-c',
        dest='config_path',
        action='store',
        metavar='<config>',
        default="./config.json",
        help='Specify path of the configuration file')
    parser.add_argument(
        '-k',
        dest='keep_temp',
        action='store_true',
        help='Keep temporary files.')
    parser.add_argument(
        '-v',
        dest='verbose',
        action='store_true',
        help='Enable verbose logging.')
    parser.add_argument(
        '-V',
        dest='very_verbose',
        action='store_true',
        help='Like -v, but more.')

    # Extra arguments for ai/inspect.py
    group_inspect = parser.add_argument_group('options for action [inspect]')
    group_inspect.add_argument(
        '-i',
        dest='mkv_path',
        metavar='<input>',
        help='Path to DTrack-Formatted MKV file (or directory of files)')

    return parser.parse_args()
