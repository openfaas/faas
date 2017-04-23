## SentimentAnalysis

Python function provides a rating on sentiment positive/negative (polarity -1.0-1.0) and subjectivity to provided to each of the sentences sent in via the [TextBlob project](http://textblob.readthedocs.io/en/dev/).

Example:

Run in the function:

```
# curl -s --fail localhost:8080/system/functions -d \
'{ 
   "service": "sentimentanalysis",
   "image": "functions/sentimentanalysis",
   "envProcess": "python ./handler.py",
   "network": "func_functions"
   }'
```

Now test the function:

```
# curl localhost:8080/function/sentimentanalysis -d "Personally I like functions to do one thing and only one thing well, it makes them more readable."
Polarity: 0.166666666667 Subjectivity: 0.6

# curl localhost:8080/function/sentimentanalysis -d "Functions are great and proven to be awesome"
Polarity: 0.9 Subjectivity: 0.875

# curl localhost:8080/function/sentimentanalysis -d "The hotel was clean, but the area was terrible"; echo
Polarity: -0.316666666667 Subjectivity: 0.85
```
