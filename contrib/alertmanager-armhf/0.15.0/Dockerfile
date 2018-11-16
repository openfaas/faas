FROM alpine:3.8
WORKDIR /root
RUN apk add --update libarchive-tools
ADD https://github.com/prometheus/alertmanager/releases/download/v0.15.0/alertmanager-0.15.0.linux-armv7.tar.gz /root/
RUN bsdtar -xvf *.tar.gz -C ./ --strip-components=1
RUN mkdir /etc/alertmanager

RUN cp alertmanager               /bin/alertmanager
RUN cp alertmanager.yml /etc/alertmanager/alertmanager.yml

EXPOSE     9093
VOLUME     [ "/alertmanager" ]
WORKDIR    /alertmanager

ENTRYPOINT [ "/bin/alertmanager" ]
CMD        [ "--config.file=/etc/alertmanager/alertmanager.yml", \
    "--storage.path=/alertmanager" ]
