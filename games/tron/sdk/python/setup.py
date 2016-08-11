#!/user/bin/env python

from setuptools import setup
from pip.req import parse_requirements

install_reqs = parse_requirements('requirements.txt', session=False)
reqs = [str(ir.req) for ir in install_reqs]

setup(name='botbox-tron',
        version='1.0',
        description='Python BotBox Tron agent SDK',
        py_modules=['botbox_tron'],
        install_requires=reqs)

