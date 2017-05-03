import sys
import json
from textblob import TextBlob

# Set encoding to UTF-8 (vs ASCII to eliminate potential errors).
reload(sys)
sys.setdefaultencoding('utf8')

def get_stdin():
    buf = ""
    for line in sys.stdin:
        buf = buf + line
    return buf

if(__name__ == "__main__"):
    st = get_stdin()
    blob = TextBlob(st)
    res = {
        "polarity": 0,
        "subjectivity": 0
    }

    for sentence in blob.sentences:
        res["subjectivity"] = res["subjectivity"] + sentence.sentiment.subjectivity
        res["polarity"] = res["polarity"] + sentence.sentiment.polarity

    total = len(blob.sentences)

    res["sentence_count"] = total
    res["polarity"] = res["polarity"] / total
    res["subjectivity"] = res["subjectivity"] / total

    print(json.dumps(res))
