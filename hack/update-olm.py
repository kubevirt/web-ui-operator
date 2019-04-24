#!/usr/bin/env python3

import logging
import sys
import yaml

_DESCRIPTION = "KubeVirt Web UI"
_ANNOTATIONS = {
    'categories': 'OpenShift Optional',
    'capabilities': 'Basic Install',
    'containerImage': 'quay.io/kubevirt/kubevirt-web-ui-operator',
    'repository': 'https://github.com/kubevirt/kubevirt-web-ui-operator',
    'createdAt': '2019-04-18T21:00:00Z',
    'description': _DESCRIPTION,
}
_SPEC = {
    'description': _DESCRIPTION,
    'provider': {
        'name': 'KubeVirt project'
    },
    'maintainers': [{
        'name': 'KubeVirt project',
        'email': 'kubevirt-dev@googlegroups.com',
    }],
    'keywords': [
        'KubeVirt', 'Virtualization', 'UI'
    ],
    'links': [{
        'name': 'KubeVirt',
        'url': 'https://kubevirt.io',
    }, {
        'name': 'Source Code',
        'url': 'https://github.com/kubevirt/web-ui-operator'
    }],
    'labels': {
        'alm-owner-kubevirt': 'kubevirt-web-ui',
        'operated-by': 'kubevirt-web-ui',
    },
    'selector': {
        'matchLabels': {
            'alm-owner-kubevirt': 'kubevirt-web-ui',
            'operated-by': 'kubevirt-web-ui',
        },
    },
}

_CRD_INFOS = {
    'kwebuis.kubevirt.io': {
        'displayName': 'KubeVirt Web UI Resource',
        'description': _DESCRIPTION,
    }
}


def process(path):
    with open(path, 'rt') as fh:
        manifest = yaml.safe_load(fh)

    manifest['metadata']['name'] = 'kubevirt-' + manifest['metadata']['name']
    manifest['metadata']['annotations'].update(_ANNOTATIONS)

    manifest['spec'].update(_SPEC)

    for crd in manifest['spec']['customresourcedefinitions']['owned']:
        crd.update(_CRD_INFOS.get(crd['name'], {}))

    yaml.safe_dump(manifest, sys.stdout)


if __name__ == '__main__':
    for arg in sys.argv[1:]:
        try:
            process(arg)
        except Exception as ex:
            logging.error('error processing %r: %s', arg, ex)
# keep going!
