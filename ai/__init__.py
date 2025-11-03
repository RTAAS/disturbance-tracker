'''
Disturbance Tracker - Machine Learning
Originally: Audio Pattern Ranger (APR)

License:   AGPL-3
Copyright: 2024, Michael Lustfield
Authors:   See history with "git log" or "git blame"
'''


def format_time(total_seconds):
    '''
    Return a formatted time (Hours:Minutes:Seconds) from seconds
    '''
    hours, remainder = divmod(int(total_seconds), 3600)
    minutes, seconds = divmod(remainder, 60)
    return f'{hours:02d}:{minutes:02d}:{seconds:02d}'
