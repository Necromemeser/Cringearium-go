"""Prepare per-service configuration, initialize databases, then exec the process supervisor."""
import os
import subprocess
import sys
import time
from pathlib import Path
from urllib.parse import quote, urlencode, urlsplit

DATABASES = {
    'auth': 'cringearium_auth',
    'courses': 'cringearium_courses',
    'ai-chat': 'cringearium_ai_chat',
    'ai-tests': 'cringearium_ai_tests',
}


def configure():
    for name in ['POSTGRES_PASSWORD', 'JWT_SECRET', 'ADMIN_PASSWORD', 'LLM_BASE_URL', 'LLM_API_KEY', 'LLM_MODEL']:
        if not os.environ.get(name, '').strip():
            raise ValueError(f'{name} is required')
    if not 8 <= len(os.environ['ADMIN_PASSWORD'].encode()) <= 72:
        raise ValueError('ADMIN_PASSWORD must contain between 8 and 72 bytes')
    if urlsplit(os.environ['LLM_BASE_URL']).scheme not in ['http', 'https']:
        raise ValueError('LLM_BASE_URL must be an HTTP(S) URL')
    port = int(os.environ.get('POSTGRES_PORT', '5432'))
    if not 1 <= port <= 65535:
        raise ValueError('POSTGRES_PORT is outside the valid range')
    host = os.environ.get('POSTGRES_HOST', 'postgres')
    user = os.environ.get('POSTGRES_USER', 'cringearium')
    password = os.environ['POSTGRES_PASSWORD']
    sslmode = os.environ.get('POSTGRES_SSLMODE', 'disable')
    if sslmode not in ['disable', 'allow', 'prefer', 'require', 'verify-ca', 'verify-full']:
        raise ValueError('Invalid POSTGRES_SSLMODE')
    os.environ.update(PGHOST=host, PGPORT=str(port), PGUSER=user, PGPASSWORD=password,
                      PGSSLMODE=sslmode, PGCONNECT_TIMEOUT='3')
    address = f'[{host}]' if ':' in host and not host.startswith('[') else host
    for service, database in DATABASES.items():
        variable = service.upper().replace('-', '_') + '_DATABASE'
        database = os.environ.get(variable, database)
        if not database or '\x00' in database or len(database.encode()) > 63:
            raise ValueError(f'{variable} must be a database name of at most 63 bytes')
        DATABASES[service] = database
        dsn = f'postgres://{quote(user, safe="")}:{quote(password, safe="")}@{address}:{port}/{quote(database, safe="")}?{urlencode({"sslmode": sslmode})}'
        os.environ[variable + '_URL'] = dsn


def prepare_databases():
    timeout = float(os.environ.get('POSTGRES_WAIT_TIMEOUT', '120'))
    if not 1 <= timeout <= 600:
        raise ValueError('POSTGRES_WAIT_TIMEOUT must be between 1 and 600 seconds')
    base = os.environ.get('POSTGRES_DB', 'postgres')
    command = ['psql', '-X', '--no-password', '-v', 'ON_ERROR_STOP=1', '-qtA', '-d', base]
    deadline = time.monotonic() + timeout
    print('Waiting for PostgreSQL...', flush=True)
    while True:
        try:
            result = subprocess.run(command + ['-c', 'SELECT 1'], stdout=subprocess.DEVNULL,
                                    stderr=subprocess.DEVNULL, timeout=5)
            if result.returncode == 0:
                break
        except subprocess.TimeoutExpired:
            pass
        if time.monotonic() >= deadline:
            raise RuntimeError('PostgreSQL is unavailable; check its address, credentials and TLS settings')
        time.sleep(1)
    # One session holds the lock while CREATE DATABASE runs outside a transaction.
    statements = ["SELECT pg_advisory_lock(hashtextextended('cringearium:database-bootstrap',0));"]
    for name in DATABASES.values():
        identifier = '"' + name.replace('"', '""') + '"'
        create = ('CREATE DATABASE ' + identifier).replace("'", "''")
        literal = name.replace("'", "''")
        statements.append(f"SELECT '{create}' WHERE NOT EXISTS(SELECT 1 FROM pg_database WHERE datname='{literal}')\\gexec")
    statements.append("SELECT pg_advisory_unlock(hashtextextended('cringearium:database-bootstrap',0));")
    result = subprocess.run(command, input='\n'.join(statements) + '\n', text=True,
                            stdout=subprocess.DEVNULL, timeout=60)
    if result.returncode:
        raise RuntimeError('Database initialization failed; create the configured databases or grant CREATEDB')
    print('PostgreSQL databases are ready.', flush=True)


def write_supervisor():
    lines = ['[supervisord]', 'nodaemon=true', 'user=root', 'logfile=/dev/null',
             'pidfile=/run/cringearium/supervisord.pid', 'childlogdir=/run/cringearium',
             '[unix_http_server]', 'file=/run/cringearium/supervisor.sock', 'chmod=0700',
             '[rpcinterface:supervisor]', 'supervisor.rpcinterface_factory=supervisor.rpcinterface:make_main_rpcinterface',
             '[supervisorctl]', 'serverurl=unix:///run/cringearium/supervisor.sock']
    programs = [('auth', '/app/bin/auth', 10), ('courses', '/app/bin/courses', 20),
                ('ai-chat', '/app/bin/ai-chat', 30), ('ai-tests', '/app/bin/ai-tests', 30),
                ('gateway', '/app/bin/gateway', 40), ('nginx', "/usr/sbin/nginx -g 'daemon off;'", 50),
                ('health', '/usr/bin/python3 /app/deploy/healthcheck.py', 60)]
    for name, command, priority in programs:
        lines += [f'[program:{name}]', f'command={command}', f'priority={priority}',
                  'directory=/app', 'autostart=true', 'autorestart=true', 'startsecs=1',
                  'startretries=30', 'stopasgroup=true', 'killasgroup=true', 'stopwaitsecs=10',
                  'stdout_logfile=/dev/stdout', 'stdout_logfile_maxbytes=0',
                  'stderr_logfile=/dev/stderr', 'stderr_logfile_maxbytes=0']
        if name in DATABASES:
            variable = name.upper().replace('-', '_') + '_DATABASE_URL'
            lines.append(f'environment=DATABASE_URL="%(ENV_{variable})s"')
        if name == 'nginx':
            lines.append('stopsignal=QUIT')
    path = Path('/run/cringearium/supervisord.conf')
    path.write_text('\n'.join(lines) + '\n')
    return str(path)


def main():
    try:
        configure()
        prepare_databases()
        config = write_supervisor()
    except (ValueError, RuntimeError, subprocess.TimeoutExpired) as error:
        print(f'Startup failed: {error}', file=sys.stderr, flush=True)
        return 1
    os.execv('/usr/bin/supervisord', ['supervisord', '-c', config])


if __name__ == '__main__':
    sys.exit(main())
