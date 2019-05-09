FROM alpine:3.9

WORKDIR /root
RUN apk add --update libarchive-tools
ADD https://github.com/prometheus/alertmanager/releases/download/v0.16.1/alertmanager-0.16.1.linux-arm64.tar.gz /root/
RUN bsdtar -xvf *.tar.gz -C ./ --strip-components=1
RUN mkdir /etc/alertmanager

RUN cp alertmanager               /bin/alertmanager
RUN cp alertmanager.yml           /etc/alertmanager/alertmanager.yml

EXPOSE     9093
VOLUME     [ "/alertmanager" ]
WORKDIR    /alertmanager

ENTRYPOINT [ "/bin/alertmanager" ]
CMD        [ "--config.file=/etc/alertmanager/alertmanager.yml", \
    "--storage.path=/alertmanager" ]
