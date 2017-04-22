import sys
from textblob import TextBlob

def get_stdin():
    buf = ""
    for line in sys.stdin:
        buf = buf + line
    return buf

if(__name__ == "__main__"):
    st = get_stdin()
    blob = TextBlob(st)
    out =""
    for sentence in blob.sentences:
        out = out + "Polarity: " + str(sentence.sentiment.polarity) + " Subjectivity: " + str(sentence.sentiment.subjectivity)  + "\n"
    print(out)

