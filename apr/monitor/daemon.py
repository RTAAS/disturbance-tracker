'''
Collect and Manage Recordings

This is ultimately a wrapper around ffmpeg that collects recordings of a
pre-determined length and performs some additional post-processing while
providing the minimum amount of blocking.
'''
import datetime
import fasteners
import logging
import pathlib
import os
import signal
import subprocess
import time

import apr.config
import apr.options
import apr.monitor.signals


def main():
    '''
    Master control logic for monitor daemon.
    '''
    workspace = pathlib.Path(apr.config.get('workspace'))
    pidfile = workspace / 'monitor.pid'
    lock = fasteners.InterProcessLock(pidfile)

    # Guard: Check for signals to be processed
    cli_signal = apr.monitor.signals.read_signal()
    if cli_signal > 0:
        # Check for PID of monitor
        if not os.path.exists(pidfile):
            logging.info('monitor.pid was not found; cannot send signal')
            return None

        running_pid = get_pid(pidfile)
        if not running_pid:
            logging.info('monitor.pid was found but not locked; deleteing')
            os.remove(pidfile)
        else:
            apr.monitor.signals.push_signal(int(running_pid), cli_signal)
            logging.debug('SENT signal %s to pid %s', cli_signal, running_pid)
            if cli_signal >= 2:
                # Wait for application to shut down
                logging.info('Waiting for monitor daemon to shut down.')
                while apr.monitor.signals.is_running(running_pid):
                    time.sleep(0.5)

        # Guard: Finished handling signals
        return None

    # Grab a process lock
    if not lock.acquire(blocking=False):
        raise Exception('Cannot grab a process lock; already running')

    # Set up signal handling
    signal.signal(signal.SIGUSR1, apr.monitor.signals.handle_signal)  # Stop
    signal.signal(signal.SIGUSR2, apr.monitor.signals.handle_signal)  # Kill

    # Enter into main processing loop
    try:
        recording_loop()
    except KeyboardInterrupt:
        pass  # Exit normally on Ctrl+C

    lock.release()


def get_pid(filename):
    '''
    Returns the PID of a given path, if it is held open.
    '''
    pidcheck = subprocess.run(
            ['lsof', '-t', filename],
            stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    return pidcheck.stdout.decode()


def recording_loop():
    '''
    Capture and process recordings until signalled to shut down.
    '''
    incoming = pathlib.Path(apr.config.get('workspace')) / 'rotating'
    record_command = build_ffmpeg_command()

    # Ensure output directory exists
    incoming.mkdir(parents=False, exist_ok=True)

    while not apr.monitor.signals.SIGTERM:
        current_time = datetime.datetime.now().strftime("%F_%H:%M:%S")
        recording = subprocess.run(
                record_command + [f'{incoming / current_time}.mkv'],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE)

        # Process any potential log messages (should be empty)
        log = recording.stderr.decode() + recording.stderr.decode()
        if log:
            logging.error('ffmpeg error:: %s', log.strip())

        # TODO: Fork a subprocess to clean up and analyze


def build_ffmpeg_command():
    '''
    Returns a list with the ffmpeg command and all required options.
    Note: This does NOT include the destination filename (last argument).
    '''
    command = [
            'ffmpeg', '-y', '-loglevel', 'error',
            '-nostdin', '-nostats']

    if apr.config.get('record_cam'):
        command.extend(['-f', 'v4l2'])
        command.extend(apr.config.get('record_cam_options'))
        command.extend([
            '-thread_queue_size', '1024',
            '-i', apr.config.get('record_cam')])

    command.extend([
        '-f', 'alsa', '-thread_queue_size', '1024',
        '-i', apr.config.get('record_mic')])

    if apr.config.get('record_cam'):
        command.extend(apr.config.get('record_cam_timestamp'))

    command.extend([
        '-preset', apr.config.get('record_compression'),
        '-t', apr.config.get('record_duration')])

    logging.debug('Constructed ffmpeg command: %s', ' '.join(command))
    return command


def delete_old():
    '''
    Remove files beyond rotation age.
    '''
    pass
    # TODO: find /home/michael/captures/ -mmin +1500 -type f -delete
