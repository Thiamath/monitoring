FROM alpine:3.9.4

RUN wget https://letsencrypt.org/certs/letsencryptauthorityx3.pem.txt -O /nalej/letsencryptauthorityx3.pem && \
	wget https://letsencrypt.org/certs/fakeleintermediatex1.pem -O /nalej/fakeleintermediatex1.pem && \
	wget https://letsencrypt.org/certs/fakelerootx1.pem -O /nalej/fakelerootx1.pem && \
	ln -s /nalej/letsencryptauthorityx3.pem /nalej/PRODUCTION.pem && \
	cat /nalej/fakeleintermediatex1.pem /nalej/fakelerootx1.pem > /nalej/STAGING.pem && \
	ln -s /nalej/STAGING.pem /nalej/DEVELOPMENT.pem

COPY monitoring-manager /nalej/

ENTRYPOINT ["./nalej/monitoring-manager"]
