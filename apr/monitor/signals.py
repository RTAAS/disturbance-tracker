'''
Collect and Manage Recordings

This is ultimately a wrapper around ffmpeg that collects recordings of a
pre-determined length and performs some additional post-processing while
providing the minimum amount of blocking.
'''
import logging
import signal
import os
import time

import apr.config
import apr.options

# Global "stop value" that can be passed between processes
SIGTERM = False
SIGKILL = False


def read_signal():
    '''
    Returns the shutdown urgency (max signal) from command line using:
    0. Continue, 1. Wait for Stop, 2. Signal to Stop, 3. Stop Immediately
    '''
    if apr.options.get('signal_halt'):
        return 3
    elif apr.options.get('signal_shutdown'):
        return 2
    elif apr.options.get('signal_stop'):
        return 1
    return 0


def push_signal(pid, sig):
    '''
    Sends a signal to the running daemon to initiate a shutdown.
    '''
    # Level of patience
    if sig in [1, 2]:
        # Send SIGUSR1 to signal graceful shutdown
        os.kill(pid, signal.SIGUSR1)
    elif sig == 3:
        # Send SIGUSR2 to signal kill
        os.kill(pid, signal.SIGUSR2)
        logging.debug('Waiting a moment before sending SIGKILL')
        time.sleep(3)
        os.kill(pid, signal.SIGKILL)


def handle_signal(signum, *args):  # pylint: disable=unused-argument
    '''
    Generic interface to handle signals.
    '''
    global SIGTERM
    global SIGKILL
    if signum == signal.SIGUSR1:
        logging.info('SIGUSR1 received; Monitor will gracefully now shut down')
        SIGTERM = True
    elif signum == signal.SIGUSR2:
        logging.warning('SIGUSR2 received; Monitor will halt immediately')
        SIGTERM = True
        SIGKILL = True


def is_running(pid):
    '''
    Returns True if a process is running.
    '''
    try:
        os.kill(int(pid), 0)
    except OSError:
        return False
    return True
