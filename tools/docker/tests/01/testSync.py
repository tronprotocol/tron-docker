import json
import sys
import time
from contextlib import closing
from http.client import HTTPConnection, HTTPException


def sync_height(timeout):
    with closing(HTTPConnection("127.0.0.1", 8090, timeout=timeout)) as node:
        node.request("GET", "/wallet/getnodeinfo")
        response = node.getresponse()
        if response.status != 200:
            raise ValueError(f"getnodeinfo returned HTTP {response.status}")
        info = json.load(response)

    height = info.get("beginSyncNum") if isinstance(info, dict) else None
    if type(height) is not int or height < 0:
        raise ValueError("getnodeinfo returned an invalid beginSyncNum")
    return height


def wait_for_sync():
    deadline = time.monotonic() + 1000
    initial = None
    while (remaining := deadline - time.monotonic()) > 0:
        try:
            height = sync_height(min(5, remaining))
        except (OSError, HTTPException, ValueError) as error:
            print(f"Retrying getnodeinfo: {error}", file=sys.stderr)
        else:
            if initial is None:
                initial = height
                print(f"Initial beginSyncNum: {initial}")
            elif height > initial:
                print(f"beginSyncNum increased from {initial} to {height}")
                return
        time.sleep(max(0, min(10, deadline - time.monotonic())))
    sys.exit("No increase in beginSyncNum observed within 1000 seconds")


if __name__ == "__main__":
    wait_for_sync()
