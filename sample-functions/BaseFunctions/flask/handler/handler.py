#!/usr/bin/python
from wsgiref.handlers import CGIHandler
from flask import render_template
from flask import Flask
import os
from flask import request

app = Flask(__name__)

@app.route('/hello/')
def hello():
    name = "test"
    return render_template('hello.html', name = name)
    # if "name" in request.args:
    #     return render_template('hello.html', name=request.args["name"])
    # else:
    #     return render_template('hello.html', name="stranger")

class ProxyFix(object):
    def __init__(self, app):
        self.app = app

    def __call__(self, environ, start_response):
        environ['SERVER_NAME'] = "localhost"
        environ['SERVER_PORT'] = "8080"
        environ['REQUEST_METHOD'] = "GET"
        environ['SCRIPT_NAME'] = ""
        environ['PATH_INFO'] = "/hello/"
        environ['QUERY_STRING'] = os.getenv("Http_Query", default="/")
        environ['SERVER_PROTOCOL'] = "HTTP/1.1"
        return self.app(environ, start_response)

if __name__ == '__main__':
    app.wsgi_app = ProxyFix(app.wsgi_app)
    CGIHandler().run(app)

