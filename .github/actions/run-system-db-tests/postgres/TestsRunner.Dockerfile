FROM golang:1.19.1
ARG POSTGRES_VERSION

ENV POSTGRES_USERNAME= \
    POSTGRES_PASSWORD= \
    POSTGRES_HOSTNAME= \
    POSTGRES_PORT= \
    POSTGRES_CERTS_PATH=/tls-certs

VOLUME /backup-and-restore-sdk-release

RUN go install github.com/onsi/ginkgo/ginkgo@latest

# https://linuxhint.com/install-postgresql-debian/
# Simply copying the binaries from postgres:${POSTGRES_VERSION} image doesn't work
# postgres binaries have many linktime and runtime dependencies which make it hard
# to keep several postgres versions installed simultaneously in different paths
RUN echo "Adding postgresql APT repository and install Postgres Client" \
 && curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /usr/share/keyrings/postgresql-keyring.gpg \
 && echo "deb [signed-by=/usr/share/keyrings/postgresql-keyring.gpg] http://apt.postgresql.org/pub/repos/apt/ bullseye-pgdg main" | tee /etc/apt/sources.list.d/postgresql.list \
 && apt-get update \
 && case "$POSTGRES_VERSION" in \
      *9.4*) \
        apt-get install postgresql-client-9.4 -y --no-install-recommends \
        ;;   \
      *9.6*) \
        apt-get install postgresql-client-9.6 -y --no-install-recommends \
        ;;   \
      *10*)  \
        apt-get install postgresql-client-10 -y --no-install-recommends \
        ;;   \
      *11*)  \
        apt-get install postgresql-client-11 -y --no-install-recommends \
        ;;   \
      *12*)  \
        apt-get install postgresql-client-12 -y --no-install-recommends \
        ;;   \
      *13*)  \
        apt-get install postgresql-client-13 -y --no-install-recommends \
        ;;   \
    esac \
 && apt-get clean

