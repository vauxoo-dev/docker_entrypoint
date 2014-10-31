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
import redis


FILESTORE_PATH = '/home/odoo/.local/share/Odoo'
CONFIGFILE_PATH = '/home/odoo/.openerp_serverrc'


def change_values(file_name, getter_func):
    '''
    Changes value from a config file

    :param str file_name: Config file name
    :param str search_str: Search sting
    :param str new_str: New string that
    '''
    for line in fileinput.input(file_name, inplace=True):
        new_str = line
        parts = line.split("=")
        if len(parts) > 1:
            search_str = parts[0].upper()
            value = getter_func(search_str)
            if value:
                new_str = "%s = %s" % (parts[0], value)
        print(new_str.replace('\n', ''))


def get_owner(file_name):
    '''
    This function gets owner name from system for a directory or file

    :param str file_name: File or directory name
    :returns: Owner name
    '''
    file_stat = stat(file_name)
    try:
        owner = pwd.getpwuid(file_stat.st_uid).pw_name
    except KeyError:
        owner = "None"
    return owner


def get_redis_vars(var_name):
    '''
    This function gets values from a has stored in redis

    :param str var_name: The key or var name
    :returns: Value
    '''
    r_server = redis.Redis(getenv('REDIS_SERVER'))
    return r_server.hget(getenv('STAGE'), var_name)


def main():
    '''
    Main entry point function
    '''

    if not path.isfile(CONFIGFILE_PATH):
        copy2("/external_files/.openerp_serverrc", CONFIGFILE_PATH)

    if getenv('REDIS_SERVER'):
        getter_func = get_redis_vars
    else:
        getter_func = getenv

    change_values(CONFIGFILE_PATH, getter_func)

    if not path.isfile(FILESTORE_PATH):
        call(["mkdir", "-p", FILESTORE_PATH])

    if get_owner(FILESTORE_PATH) != "odoo":
        call(["chown", "-R", "odoo:odoo", "/home/odoo"])

    call(["/usr/bin/supervisord"])


if __name__ == '__main__':
    main()
