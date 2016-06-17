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
import logging
import sys
import traceback
import string
import random

logging.basicConfig(
    level=logging.DEBUG,
    format='%(levelname)s:[%(asctime)s] - %(name)s.%(module)s - %(message)s')
logger = logging.getLogger("entry_point")

USER_NAME = getenv('ODOO_USER') and getenv('ODOO_USER') or 'odoo'

FILESTORE_PATH = getenv('ODOO_FILESTORE_PATH') \
    and getenv('ODOO_FILESTORE_PATH') \
    or '/home/%s/.local/share/Odoo/filestore' % USER_NAME

CONFIGFILE_PATH = getenv('ODOO_CONFIG_FILE') \
    and getenv('ODOO_CONFIG_FILE') \
    or '/home/%s/.openerp_serverrc' % USER_NAME


def change_values(file_name, getter_func):
    '''
    Changes value from a config file, new values are gotten from redis server
    or env vars

    :param str file_name: Config file name
    :getter_func: Function that will be used for getting new values
    '''
    for line in fileinput.input(file_name, inplace=True):
        new_str = line
        logger.debug("Line readed: %s", line.strip())
        parts = line.split("=")
        logger.debug("Parts: %s", len(parts))
        if len(parts) > 1:
            search_str = parts[0].upper().strip()
            value = getter_func(search_str)
            logger.debug("Search for: %s and value is: %s", search_str, value)
            if search_str == 'ADMIN_PASSWD' and \
                (not value or value == 'admin'):
                value = ''.join(random.choice(string.letters+string.digits) for i in range(12))
            if value:
                new_str = "%s = %s" % (parts[0].strip(), value.strip())
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
    logger.debug("Owner of %s is %s", file_name, owner)
    return owner


def main():
    '''
    Main entry point function
    '''
    logger.info("Entering entry point main function")
    if not path.isfile(CONFIGFILE_PATH):
        copy2("/external_files/openerp_serverrc", CONFIGFILE_PATH)

    getter_func = getenv
    logger.info("Using env vars")

    change_values(CONFIGFILE_PATH, getter_func)

    if not path.exists(FILESTORE_PATH):
        call(["mkdir", "-p", FILESTORE_PATH])

    logger.info("Setting permissions")
    logger.debug("Changing /tmp to 'ugo+rwxt'")
    call(["chmod", "ugo+rwxt", "/tmp"])
    logger.debug("Changing '/home/%s' owner to '%s:%s'", USER_NAME, USER_NAME, USER_NAME)
    call(["chown", "-R", "%s:%s" % (USER_NAME, USER_NAME),
          "/home/%s" % USER_NAME])
    logger.info("All changes made, now will run supervisord")
    call(["supervisord", "-c", "/etc/supervisor/supervisord.conf"])


if __name__ == '__main__':
    main()
