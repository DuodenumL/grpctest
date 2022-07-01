from __future__ import annotations

import json
import operator
import os.path
import typing

import sys
import yaml


def itemsetter(obj, key):
    return lambda v: operator.setitem(obj, key, v)


def need_combine(key: str) -> bool:
    return key.endswith('@')


def traverse_combinators(
        obj: typing.Union[dict, list],
) -> list[tuple[list[typing.Any], callable[[typing.Any]]]]:
    ret = []

    if isinstance(obj, dict):
        keys_to_rename = []
        for key, value in obj.items():
            if need_combine(key):
                keys_to_rename.append(key)
            else:
                ret.extend(traverse_combinators(value))

        for old_key in keys_to_rename:
            new_key = old_key.rstrip('@')
            value = obj[new_key] = obj[old_key]
            ret.append((value, itemsetter(obj, new_key)))
            del obj[old_key]

    elif isinstance(obj, list):
        for value in obj:
            ret.extend(traverse_combinators(value))

    return ret


def combine_requests(req: dict) -> typing.Generator[dict, None, None]:
    combinators = traverse_combinators(req)

    def combine_req_for_one(idx: int) -> typing.Generator[dict, None, None]:
        if idx == len(combinators):
            yield req
            return

        combine_args, setter = combinators[idx]
        for arg in combine_args:
            setter(arg)
            yield from combine_req_for_one(idx + 1)

    yield from combine_req_for_one(0)


def main():
    rpcs = []
    path = sys.argv[1]
    if os.path.isdir(path):
        for filename in os.listdir(path):
            if filename.endswith('.yaml'):
                with open(os.path.join(path, filename)) as f:
                    rpcs.extend(yaml.safe_load(f))
    else:
        with open(path) as f:
            rpcs.extend(yaml.safe_load(f))

    for rpc in rpcs:
        for req in combine_requests(rpc['requests']):
            print(rpc['method'])
            print(json.dumps(rpc.get('prepare', '')))
            print(json.dumps(req))
            print(json.dumps(rpc.get('asserts', [])))


if __name__ == '__main__':
    main()
