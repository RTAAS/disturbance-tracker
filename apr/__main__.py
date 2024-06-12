#!/usr/bin/env python3
'''
Primary entry point
'''
import os
import importlib

import apr.config
import apr.options


def main():
    '''
    Prepare and launch application.
    '''
    # Use config from command line
    config_path = apr.options.get('config_path')
    if config_path:
        os.environ['APR_CONFIG'] = str(config_path)
    # Load application configuration
    apr.config.load_configuration()

    # Kick off main process
    selected_action = apr.options.get('action')
    if not selected_action:
        raise Exception('No action (-a) provided!')
    router = importlib.import_module(f'apr.{selected_action}')
    if not router:
        raise Exception('Unable to load selected action!')
    router.entry_point()


if __name__ == '__main__':
    main()
