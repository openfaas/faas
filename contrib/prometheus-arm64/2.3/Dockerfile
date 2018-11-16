FROM alpine:3.8
WORKDIR /root/

RUN apk add --update libarchive-tools curl \
    && curl -sLf https://github.com/prometheus/prometheus/releases/download/v2.3.1/prometheus-2.3.1.linux-arm64.tar.gz > prometheus.tar.gz \
    && bsdtar -xvf prometheus.tar.gz -C ./ --strip-components=1 \
    && apk del libarchive-tools curl \
    && mkdir /etc/prometheus \
    && mkdir -p /usr/share/prometheus \
    && cp prometheus                             /bin/prometheus \
    && cp promtool                               /bin/promtool \
    && cp prometheus.yml  				/etc/prometheus/ \
    && cp -r console_libraries                     /usr/share/prometheus/ \
    && cp -r consoles                              /usr/share/prometheus/ \
    && rm -rf /root/*

RUN ln -s /usr/share/prometheus/console_libraries /usr/share/prometheus/consoles/ /etc/prometheus/
RUN mkdir -p /prometheus && \
    chown -R nobody:nogroup /etc/prometheus /prometheus

USER       nobody
EXPOSE     9090
VOLUME     [ "/prometheus" ]
WORKDIR    /prometheus

ENTRYPOINT [ "/bin/prometheus" ]
CMD        [ "--config.file=/etc/prometheus/prometheus.yml", \
    "--storage.tsdb.path=/prometheus", \
    "--web.console.libraries=/usr/share/prometheus/console_libraries", \
    "--web.console.templates=/usr/share/prometheus/consoles" ]

