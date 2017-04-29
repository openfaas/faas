# Azure Template Parameter Packer
# Takes username as first argument and outputs parameter JSON on stdout

from base64 import b64encode
from os import environ
from json import dumps
from os.path import exists, join
from sys import argv, stderr, stdout

USERNAME = argv[1].strip()
OWN_PUBKEY = join(environ['HOME'],'.ssh','id_rsa.pub')
GEN_PUBKEY = USERNAME + '.pub'

if exists(GEN_PUBKEY):
    admin_public_key = open(GEN_PUBKEY,'r').read()
elif exists(OWN_PUBKEY):
    admin_public_key = open(OWN_PUBKEY,'r').read()
    stderr.write('Warning: using %s instead of freshly generated keys.\n' % OWN_PUBKEY)
else:
    stderr.write('No public keys found, exiting.\n')
    exit(1)

params = {
    "adminUsername": {
        "value": USERNAME
    },
    "adminPublicKey": {
        "value": admin_public_key
    },
    "masterCount": { 
        "value": int(environ.get('MASTER_COUNT', 1))
    },
    "masterCustomData": {
        "value": b64encode(open("cloud-config-master.yml", "r").read())
    },
    "agentCount": {
        "value": int(environ.get('AGENT_COUNT', 2))
    },
    "agentCustomData": {
        "value": b64encode(open("cloud-config-agent.yml", "r").read())
    },
    "masterSize": {
        "value": "Standard_F1"
    },
    "agentSize": {
        "value": "Standard_F1"
    },
    "saType": {
        "value": "Standard_LRS"
    }
}

stdout.write(dumps(params))