gif-maker
=========

Converts a .mov QuickTime video to a .gif

Testing:

```
$ docker build -t alexellis/gif-maker .
$ faas-cli deploy --fprocess="./entry.sh" \
  --env read_timeout=60 --env write_timeout=60 \
  --image alexellis/gif-maker --name gif-maker

# wait a little

$ curl http://localhost:8080/function/gif-maker --data-binary @$HOME/Desktop/screen1.mov > screen1.gif
```

Try to use a small cropped video around 5MB. Timeouts may need to be extended for larger videos

