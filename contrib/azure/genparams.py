from base64 import b64encode
from os import environ
from json import dumps
from os.path import exists, join

OWN_PUBKEY = join(environ['HOME'],'.ssh','id_rsa.pub')
GEN_PUBKEY = 'faas.pub'

if exists(GEN_PUBKEY):
    admin_public_key = open(GEN_PUBKEY,'r').read()
elif exists(OWN_PUBKEY):
    admin_public_key = open(OWN_PUBKEY,'r').read()
else:
    print('No public keys found, exiting.')
    exit(1)

params = {
    "adminUsername": {
        "value": "faas" # if you want to customize this, update the yml files too 
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

with open('faas-cluster-parameters.json', 'w') as h:
    h.write(dumps(params))