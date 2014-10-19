#!/usr/bin/python
# -*- coding: utf-8 -*-
'''
Entry point for Dockerized aplications, this works mainly with
Odoo instances that will be launched using supervisor
'''
from os import stat, path, getenv
from subprocess import call
from shutil import copy2
import pwd
import fileinput

CONFIGFILE_PATH = '/home/odoo/instance/config/odoo.conf'


def change_value(file_name, search_str, new_str):
    '''
    Changes value from a config file

    :param str file_name: Config file name
    :param str search_str: Search sting
    :param str new_str: New string that
    '''
    for line in fileinput.input(file_name, inplace=True):
        if line.startswith(search_str):
            print(new_str)
        else:
            print(line.replace('\n', ''))


def get_owner(file_name):
    '''
    This function gets owner name from system for a directory or file

    :param str file_name: File or directory name
    :returns: Owner name
    '''
    file_stat = stat(file_name)
    return pwd.getpwuid(file_stat.st_uid).pw_name

    call(["/usr/bin/supervisord"])

if __name__ == '__main__':
    main()
