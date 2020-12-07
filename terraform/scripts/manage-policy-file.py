#!/usr/bin/env python3
import argparse
import copy
import json
from typing import Any, Dict, Tuple

BINDINGS = 'bindings'
ROLE = 'role'
MEMBERS = 'members'

ADD = 'add'
REMOVE = 'remove'


def main(in_file: str,
         out_file: str,
         action: str,
         email: str,
         typ: str,
         role: str) -> int:
    with open(in_file, 'r') as f:
        try:
            data = json.loads(f.read())
        except json.decoder.JSONDecodeError:
            print(f'{in_file} does not contain valid json')
            return 1

    name = f'{typ}:{email}'
    if action == ADD:
        updated, err = add_user(data, name, role)
    else:
        assert action == REMOVE
        updated, err = remove_user(data, name, role)
    if err:
        print(f'could not perform update: {err}')
        return 1

    out = out_file if out_file else in_file
    with open(out, 'w') as f:
        s = json.dumps(updated, indent=4)
        f.write(s)
        print(s)

    return 0


def add_user(data: Dict[str, Any],
             name: str,
             role: str) -> Tuple[Dict[str, Any], str]:
    d = copy.deepcopy(data)
    bindings = d.get(BINDINGS, [])
    if not bindings:
        d[BINDINGS] = [
            {ROLE: role, MEMBERS: [name]}
        ]
        return d, ''

    bs = []
    found = False
    for b in bindings:
        if b[ROLE] == role:
            found = True
            if name in b[MEMBERS]:
                # No updates needed
                return d, 'user is already bound to this role'
            else:
                b[MEMBERS].append(name)
        bs.append(b)
    if not found:
        bs.append({ROLE: role, MEMBERS: [name]})

    d[BINDINGS] = bs
    return d, ''


def remove_user(data: Dict[str, Any],
                name: str,
                role: str) -> Tuple[Dict[str, Any], str]:
    d = copy.deepcopy(data)
    bindings = d.get(BINDINGS, [])
    if not bindings:
        return d, 'policy does not have bindings to remove anyone from'
    bs = []
    for b in bindings:
        members = b[MEMBERS]
        if b[ROLE] == role:
            if name not in members:
                return d, 'user is not bound to this role'
            else:
                members = [m for m in b[MEMBERS] if m != name]
        if members:
            bs.append(b)

    if not bs:
        del d[BINDINGS]
    else:
        d[BINDINGS] = bs
    return d, ''


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='foo')

    parser.add_argument('-i', '--input',
                        help='input policy file (must be json)',
                        required=True)
    parser.add_argument('-o', '--output',
                        help='output policy file. If nog specified, the input '
                             'file is overwritten.',
                        default='')
    parser.add_argument('-a', '--action',
                        help='action to perform',
                        choices=[ADD, REMOVE],
                        required=True)
    parser.add_argument('-e', '--email',
                        help='email of user/service account',
                        required=True)
    parser.add_argument('-t', '--type',
                        help='type of user',
                        choices=['user', 'serviceAccount'],
                        default='serviceAccount')
    parser.add_argument('-r', '--role',
                        help='role to edit',
                        required=True)

    args = parser.parse_args()

    exit(main(args.input, args.output, args.action, args.email, args.type,
              args.role))

