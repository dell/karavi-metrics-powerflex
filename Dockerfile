FROM scratch
ARG SERVICE
COPY $SERVICE/bin/service /service
ENTRYPOINT ["/service"]