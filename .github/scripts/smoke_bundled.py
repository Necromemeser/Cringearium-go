"""Exercise a disposable two-container deployment, including SSE with a local fake LLM.
Requires SMOKE_BASE_URL and Docker access. Never use this test against production.
"""
import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.request
import uuid

sys.dont_write_bytecode = True
os.environ.setdefault('SMOKE_BASE_URL', 'http://localhost:3000')
os.environ.setdefault('SMOKE_POSTGRES_CONTAINER', 'cringearium-bundled-postgres')
from smoke_backend import BASE_URL, request
from research_api_smoke import auth_sql

APP = os.environ.get('SMOKE_APP_CONTAINER', 'cringearium-bundled-app')
MOCK = r'''
import json, threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
release = threading.Event()
class Mock(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/release': release.set()
        self.send_response(200); self.end_headers()
    def do_POST(self):
        body=json.loads(self.rfile.read(int(self.headers['Content-Length'])))
        self.send_response(200)
        if body.get('stream'):
            self.send_header('Content-Type','text/event-stream'); self.end_headers()
            self.wfile.write(('data: '+json.dumps({'choices':[{'delta':{'content':'Тестовый '}}]})+'\n\n').encode()); self.wfile.flush()
            release.wait(15)
            self.wfile.write(('data: '+json.dumps({'choices':[{'delta':{'content':'ответ'}}]})+'\n\ndata: [DONE]\n\n').encode()); self.wfile.flush()
        else:
            self.send_header('Content-Type','application/json'); self.end_headers()
            self.wfile.write(json.dumps({'choices':[{'message':{'content':'Учебная проверка'}}]}).encode())
    def log_message(self,*args): pass
ThreadingHTTPServer(('127.0.0.1',9999),Mock).serve_forever()
'''


def container(*args, input=None):
    return subprocess.run(['docker', 'exec', '-i', APP, *args], input=input, text=True,
                          capture_output=True, check=True).stdout.strip()


def wait_ready(timeout=240):
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        try:
            with urllib.request.urlopen(BASE_URL + '/healthz', timeout=10) as response:
                if response.status == 200:
                    return
        except (OSError, urllib.error.HTTPError):
            pass
        time.sleep(1)
    raise RuntimeError('Bundled application did not become ready')


def main():
    wait_ready()
    with urllib.request.urlopen(BASE_URL + '/courses/1', timeout=10) as response:
        assert b'<html' in response.read().lower(), 'SPA fallback is broken'
    with urllib.request.urlopen(BASE_URL + '/images/memes/doge.jpg', timeout=10) as response:
        assert response.headers.get_content_type() == 'image/jpeg'
    assert 'syntax is ok' in subprocess.run(['docker','exec',APP,'nginx','-t'],capture_output=True,text=True,check=True).stderr
    runtime = json.loads(subprocess.check_output(['docker','inspect',APP],text=True))[0]
    assert set(runtime['NetworkSettings']['Ports']) == {'80/tcp'}, 'Internal service ports were published'
    admin_password = container('printenv', 'ADMIN_PASSWORD')
    admin = request('POST','/api/auth/login',{'email':'admin@mail.ru','password':admin_password})['token']
    users = request('GET','/api/admin/users',token=admin)
    assert all(any(u['username']==f'user{i}' and u['role']=='student' for u in users) for i in range(1,51))
    for i in [1,50]:
        assert request('POST','/api/auth/login',{'email':f'user{i}@mail.ru','password':'12345678'})['token']
    assert container('printenv','LLM_BASE_URL') in ['http://localhost:9999','http://127.0.0.1:9999'], 'SSE test requires the local mock provider'
    container('sh','-c','cat > /run/cringearium/mock_llm.py',input=MOCK)
    subprocess.run(['docker','exec','-d',APP,'python3','/run/cringearium/mock_llm.py'],check=True)
    time.sleep(.5)
    name='bundle_'+uuid.uuid4().hex
    credentials={'email':name+'@example.test','password':uuid.uuid4().hex}
    user=request('POST','/api/auth/register',{'username':name,**credentials},expected=201)
    token=request('POST','/api/auth/login',credentials)['token']
    chat=None
    try:
        chat=request('POST','/api/chats',{},token,201)['id']
        assert request('POST','/api/chats',{},token,201)['id']==chat
        req=urllib.request.Request(BASE_URL+f'/api/chats/{chat}/messages',data=json.dumps({'content':'Проверка потокового ответа'}).encode(),headers={'Content-Type':'application/json','Authorization':'Bearer '+token},method='POST')
        started=time.monotonic()
        with urllib.request.urlopen(req,timeout=20) as response:
            assert response.headers.get_content_type()=='text/event-stream'
            first=response.readline().decode()
            while not first.startswith('data:'):
                if not first: raise AssertionError('SSE stream ended before its first event')
                first=response.readline().decode()
            event=json.loads(first[5:].strip())
            assert event['type']=='chunk' and event['content']=='Тестовый '
            assert time.monotonic()-started<10, 'First SSE chunk was buffered until the provider completed'
            container('python3','-c',"import urllib.request; urllib.request.urlopen('http://127.0.0.1:9999/release').read()")
            events=[event]+[json.loads(line.decode()[5:].strip()) for line in response if line.startswith(b'data:')]
        assert any(e['type']=='done' for e in events) and not any(e['type']=='error' for e in events)
        assert ''.join(e['content'] for e in events if e['type']=='chunk')=='Тестовый ответ'
        messages=request('GET',f'/api/chats/{chat}/messages',token=token)
        assert len(messages)==2 and messages[-1]['content']=='Тестовый ответ'
        # Readiness must account for every service, not just the gateway's own health endpoint.
        container('supervisorctl','-c','/run/cringearium/supervisord.conf','stop','auth')
        time.sleep(2.1)
        try:
            urllib.request.urlopen(BASE_URL+'/healthz',timeout=10)
            raise AssertionError('Readiness ignored a stopped auth service')
        except urllib.error.HTTPError as error:
            assert error.code==503
        container('supervisorctl','-c','/run/cringearium/supervisord.conf','start','auth')
        wait_ready()
        previous=container('supervisorctl','-c','/run/cringearium/supervisord.conf','pid','courses')
        container('python3','-c','import os, signal, sys; os.kill(int(sys.argv[1]), signal.SIGTERM)',previous)
        deadline=time.monotonic()+30
        while time.monotonic()<deadline:
            current=container('supervisorctl','-c','/run/cringearium/supervisord.conf','pid','courses')
            if current not in ['0',previous]: break
            time.sleep(.5)
        else: raise AssertionError('Supervisor did not restart courses')
        wait_ready()
        subprocess.run(['docker','restart','--time','60',APP],check=True,capture_output=True)
        wait_ready()
        assert request('GET',f'/api/chats/{chat}/messages',token=token)==messages, 'Restart lost data or invalidated an unchanged JWT secret'
    finally:
        if chat is not None: request('DELETE',f'/api/chats/{chat}',token=token,expected=204)
        auth_sql(f"DELETE FROM users WHERE id={int(user['id'])};")
    print('PASS bundled deployment: frontend/assets, admin/students, private ports, chat reuse, live SSE, aggregated readiness, process recovery, restart persistence')


if __name__=='__main__': main()
