#!/usr/bin/env python3
import subprocess
import os
import sys

def main():
	if len(sys.argv) < 2:
		print('Please provide an op-geth commit or branch name')
		sys.exit(1)

	version = sys.argv[1]
	for project in ('.',):
		print(f'Updating {project}...')
		update_mod(project, version)


def update_mod(project, version):
	print('Replacing...')
	subprocess.run([
		'go',
		'mod',
		'edit',
		'-replace',
		f'github.com/ethereum/go-ethereum=github.com/celo-org/op-geth@{version}'
	], cwd=os.path.join(project), check=True)
	print('Tidying...')
	subprocess.run([
		'go',
		'mod',
		'tidy'
	], cwd=os.path.join(project), check=True)


if __name__ == '__main__':
	main()
