#!/usr/bin/python
from flask import render_template, Flask, request, redirect

app = Flask(__name__)


@app.route('/')
def hello():
    if "nickname" in request.args and "name" in request.args:
        return render_template('hello.html', nickname=request.args["nickname"], name=request.args["name"])
    elif "nickname" in request.args:
        return render_template('hello.html', nickname=request.args["nickname"])
    elif "name" in request.args:
        return render_template('hello.html', name=request.args["name"])
    else:
        return render_template('hello.html', name="stranger")


@app.route('/test')
def test():
    if "name" in request.args:
        return render_template('hello.html', name=request.args["name"])
    else:
        return render_template('hello.html', name="testy McTest Face")

if __name__ == '__main__':
    app.run(debug=True)
