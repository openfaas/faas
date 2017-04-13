FROM docker:dind
WORKDIR /root/
RUN apk --no-cache add git && \
  git clone https://github.com/alexellis/faas
WORKDIR /root/faas/


EXPOSE 8080
EXPOSE 9090

COPY entry.sh .
RUN chmod +x ./entry.sh

CMD ["./entry.sh"]
