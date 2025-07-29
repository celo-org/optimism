#!/usr/bin/env python3
"""
Port forwarder for Kurtosis devnet L2 RPC.
Forwards traffic from a fixed port (9545) to the dynamic L2 RPC port.
"""

import subprocess
import socket
import threading
import sys
import signal
import os

def get_l2_rpc_port():
    """Get the L2 RPC port from Kurtosis devnet"""
    try:
        result = subprocess.run([
            'kurtosis', 'port', 'print', 'celo-isthmus-devnet',
            'op-el-1-op-geth-op-node-op-kurtosis', 'rpc'
        ], capture_output=True, text=True, check=True)

        # Extract port from output like "http://127.0.0.1:32891"
        return int(result.stdout.strip().split(':')[-1])
    except subprocess.CalledProcessError as e:
        print(f"Error getting port: {e}")
        sys.exit(1)

def forward_connection(client_socket, target_socket):
    """Forward data between client and target sockets"""
    try:
        while True:
            data = client_socket.recv(4096)
            if not data:
                break
            target_socket.send(data)
    except:
        pass
    finally:
        client_socket.close()
        target_socket.close()

def forward_port(source_port, target_port):
    """Start port forwarding from source_port to target_port"""
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind(('127.0.0.1', source_port))
    server.listen(5)

    print(f"Port forwarding: 127.0.0.1:{source_port} -> 127.0.0.1:{target_port}")

    try:
        while True:
            client_socket, addr = server.accept()
            print(f"Connection from {addr}")

            # Connect to target
            target_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            target_socket.connect(('127.0.0.1', target_port))

            # Start bidirectional forwarding
            t1 = threading.Thread(target=forward_connection, args=(client_socket, target_socket))
            t2 = threading.Thread(target=forward_connection, args=(target_socket, client_socket))
            t1.daemon = t2.daemon = True
            t1.start()
            t2.start()

    except KeyboardInterrupt:
        print("\nShutting down...")
    finally:
        server.close()

def main():
    l2_rpc_port = get_l2_rpc_port()
    forward_port(9545, l2_rpc_port)

if __name__ == "__main__":
    main()
