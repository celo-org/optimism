#!/usr/bin/env python3
"""Set the Semgrep baseline from GitHub's pull request merge ref."""

import os
from pathlib import Path
import re
import subprocess


def configure_baseline():
    env_file = Path(os.environ["BASH_ENV"])
    pull_request = os.environ.get("CIRCLE_PULL_REQUEST")
    if not pull_request:
        with env_file.open("a") as output:
            output.write("unset SEMGREP_BASELINE_REF\n")
        return

    number = pull_request.rsplit("/", 1)[-1]
    if not re.fullmatch(r"[1-9][0-9]*", number):
        raise ValueError("Invalid pull request number")

    fetch = ["git", "fetch", "--no-tags"]
    shallow = subprocess.check_output(
        ["git", "rev-parse", "--is-shallow-repository"], text=True
    ).strip()
    if shallow == "true":
        fetch.append("--unshallow")
    subprocess.run([*fetch, "origin", f"refs/pull/{number}/merge"], check=True)
    merge = subprocess.check_output(
        ["git", "rev-list", "--parents", "-n", "1", "FETCH_HEAD"], text=True
    ).split()
    if len(merge) != 3:
        raise ValueError("PR merge ref must have exactly two parents")
    checkout = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
    if merge[2] != checkout:
        raise ValueError("PR merge ref does not match the checked-out PR head")

    base = merge[1]
    common = subprocess.check_output(["git", "merge-base", "HEAD", base], text=True).strip()
    with env_file.open("a") as output:
        output.write(f"export SEMGREP_BASELINE_REF={common}\n")
    print(f"Semgrep baseline: {common} (PR #{number}, base {base})")


if __name__ == "__main__":
    configure_baseline()
