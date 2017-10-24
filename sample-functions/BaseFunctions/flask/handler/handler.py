#!/usr/bin/python
from wsgiref.handlers import CGIHandler
import os
from app import app
import time

query_params = os.getenv("Http_Query", default="")
whole_path = os.getenv("Http_Path", default="")
split_path = whole_path.split('/')
if len(split_path) > 3:
    route_path = '/' + '/'.join(split_path[3:])
else:
    route_path = "/"
http_method = os.getenv("Http_Method", default="GET")


class ProxyFix(object):
    def __init__(self, app):
        self.app = app

    def __call__(self, environ, start_response):
        environ['SERVER_NAME'] = "localhost"
        environ['SERVER_PORT'] = "8080"
        environ['REQUEST_METHOD'] = http_method
        environ['SCRIPT_NAME'] = ""
        environ['PATH_INFO'] = route_path
        environ['QUERY_STRING'] = query_params
        environ['SERVER_PROTOCOL'] = "HTTP/1.1"
        return self.app(environ, start_response)

if __name__ == '__main__':
    app.wsgi_app = ProxyFix(app.wsgi_app)
    CGIHandler().run(app)


