'''
Command Line Options
'''
import argparse

# Container for parsed command line options
loaded_options = None


def get(option):
    '''
    Return the value of a command line argument.
    '''
    global loaded_options
    if not loaded_options:
        loaded_options = parser().parse_args()
    return getattr(loaded_options, option)


def parser():
    '''
    Returns an argument parser with available options.
    '''
    parser = argparse.ArgumentParser(
            usage='apr [-h] -a <action> [other_options]',
            formatter_class=lambda prog: argparse.HelpFormatter(
                prog, max_help_position=30))

    # Basic Runtime Options
    parser.add_argument(
        '-c',
        dest='config_path',
        action='store',
        metavar='<path>',
        help='Specify path of the configuration file')
    parser.add_argument(
        '-a',
        dest='action',
        metavar='<action>',
        choices=['monitor', 'train', 'inspect'],
        help='Action to perform (Available: monitor, train, inspect)')

    # Options for Monitor
    monitor = parser.add_argument_group('monitor')
    monitor.add_argument(
        '-s',
        dest='signal_stop',
        action='store_true',
        help='Stop a running daemon after current capture completes')
    monitor.add_argument(
        '-S',
        dest='signal_shutdown',
        action='store_true',
        help='(uppercase S); same as -s but waits for shutdown')
    monitor.add_argument(
        '-H',
        dest='signal_halt',
        action='store_true',
        help='Stop a running daemon immediately; will corrupt video')

    return parser
