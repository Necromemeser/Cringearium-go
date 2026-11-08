"""Readiness for all local application services and PostgreSQL. Never calls the external LLM."""
import os
import subprocess
import sys
import time
import urllib.request
from http.server import BaseHTTPRequestHandler, HTTPServer

LOCAL_HTTP = urllib.request.build_opener(urllib.request.ProxyHandler({}))


def ready():
    try:
        # Parallel checks keep health probes bounded even when several children are unavailable.
        from concurrent.futures import ThreadPoolExecutor
        def service(port):
            try:
                with LOCAL_HTTP.open(f'http://127.0.0.1:{port}/health', timeout=1) as response:
                    return response.status == 200
            except Exception:
                return False
        with ThreadPoolExecutor(max_workers=5) as executor:
            services = all(executor.map(service, [8080, 8081, 8082, 8084, 8085]))
        if not services:
            return False
        result = subprocess.run(['psql', '-X', '--no-password', '-qtA', '-d', os.environ.get('POSTGRES_DB', 'postgres'), '-c', 'SELECT 1'],
                                stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=4)
        return result.returncode == 0
    except (OSError, subprocess.TimeoutExpired):
        return False


class Handler(BaseHTTPRequestHandler):
    last_check = 0.0
    last_result = False

    def do_GET(self):
        if self.path != '/healthz':
            self.send_error(404)
            return
        if time.monotonic() - Handler.last_check >= 2:
            Handler.last_result = ready()
            Handler.last_check = time.monotonic()
        body = b'ok\n' if Handler.last_result else b'not ready\n'
        self.send_response(200 if Handler.last_result else 503)
        self.send_header('Content-Type', 'text/plain; charset=utf-8')
        self.send_header('Cache-Control', 'no-store')
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *_):
        pass


if __name__ == '__main__':
    if '--check' in sys.argv:
        # Exercise Nginx and the aggregated readiness endpoint too.
        try:
            with LOCAL_HTTP.open('http://127.0.0.1/healthz', timeout=8) as response:
                sys.exit(0 if response.status == 200 else 1)
        except Exception:
            sys.exit(1)
    HTTPServer(('127.0.0.1', 8090), Handler).serve_forever()
