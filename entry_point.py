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

FILESTORE_PATH = '/home/odoo/instance/odoo/openerp/filestore'
CONFIGFILE_PATH = '/home/odoo/.openerp_serverrc'


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


def main():
    '''
    Main entry point function
    '''

    if not path.isfile(CONFIGFILE_PATH):
        copy2("/external_files/odoo.conf", CONFIGFILE_PATH)

    if getenv('DB_SERVER'):
        change_value(CONFIGFILE_PATH, 'db_host', 'db_host = %s' % getenv('DB_SERVER'))

    if getenv('DB_PORT'):
        change_value(CONFIGFILE_PATH, 'db_port', 'db_port = %s' % getenv('DB_PORT'))

    if get_owner(CONFIGFILE_PATH) != "odoo":
        call(["chown", "-R", "odoo", CONFIGFILE_PATH])

    if get_owner(FILESTORE_PATH) != "odoo":
        call(["chown", "-R", "odoo", FILESTORE_PATH])


if __name__ == '__main__':
    main()
