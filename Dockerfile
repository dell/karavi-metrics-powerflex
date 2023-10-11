FROM registry.access.redhat.com/ubi9/ubi-micro@sha256:21daf4c8bea788f6114822ab2d4a23cca6c682bdccc8aa7cae1124bcd8002066
LABEL vendor="Dell Inc." \
      name="csm-metrics-powerflex" \
      summary="Dell Container Storage Modules (CSM) for Observability - Metrics for PowerFlex" \
      description="Provides insight into storage usage and performance as it relates to the CSI (Container Storage Interface) Driver for Dell PowerFlex" \
      version="2.0.0" \
      license="Apache-2.0"
ARG SERVICE
COPY $SERVICE/bin/service /service
ENTRYPOINT ["/service"]
