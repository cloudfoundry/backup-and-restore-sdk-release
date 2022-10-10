#########################################################################
# Find, download and uncompress latest version of $STEMCELL_NAME from Bosh.io
#########################################################################
FROM alpine as uncompressed-stemcell
ARG STEMCELL_NAME
RUN apk add curl --no-cache
RUN STEMCELL_URL="$(curl -L https://bosh.io/stemcells | grep -io "https:\/\/.*warden-boshlite-${STEMCELL_NAME}-go_agent.tgz")" \
 && curl -o /stemcell.tgz -L ${STEMCELL_URL}
RUN mkdir -p /wrapper && tar xf /stemcell.tgz -C /wrapper
RUN mkdir -p /stemcell && tar xf /wrapper/image -C /stemcell

#########################################################################
# Run bosh create-release and uncompress tarball which contains blobs and vendored packages
#########################################################################
FROM alpine as uncompressed-release
RUN apk add bash curl git ca-certificates --no-cache
RUN curl -o /bin/bosh -L "https://github.com/cloudfoundry/bosh-cli/releases/download/v7.0.1/bosh-cli-7.0.1-linux-amd64" && chmod +x /bin/bosh

COPY . /release-repo
WORKDIR /release-repo

RUN bosh create-release --force --tarball /release.tgz
RUN mkdir -p /release && tar -xf /release.tgz -C /release

WORKDIR /release/packages
RUN bash -c 'for tgz in *.tgz ; do    \
  filename=$(basename -- "${tgz}") ;  \
  ext="${filename##*.}"            ;  \
  noext="${filename%.*}"           ;  \
                                      \
  mkdir -p "${noext}"              ;  \
  tar -xf "${tgz}" -C "${noext}"   ;  \
  rm "${tgz}"                      ;  \
  echo "${tgz} extracted!"         ;  \
done'

WORKDIR /release/jobs
RUN bash -c 'for tgz in *.tgz ; do    \
  filename=$(basename -- "${tgz}") ;  \
  ext="${filename##*.}"            ;  \
  noext="${filename%.*}"           ;  \
                                      \
  mkdir -p "${noext}"              ;  \
  tar -xf "${tgz}" -C "${noext}"   ;  \
  rm "${tgz}"                      ;  \
  echo "${tgz} extracted!"         ;  \
done'

#########################################################################
# Mount alongside stemcell and uncompressed release files and run packaging scripts
#########################################################################
FROM scratch as compiled-release
COPY --from=uncompressed-stemcell /stemcell/ /
COPY --from=uncompressed-release /release /release

ENV PACKAGES="libpcre2                \
libopenssl1                           \
database-backup-restorer-mariadb      \
database-backup-restorer-boost        \
database-backup-restorer-mysql-5.6    \
database-backup-restorer-mysql-5.7    \
database-backup-restorer-mysql-8.0    \
database-backup-restorer-postgres-9.4 \
database-backup-restorer-postgres-9.6 \
database-backup-restorer-postgres-10  \
database-backup-restorer-postgres-10  \
database-backup-restorer-postgres-11  \
database-backup-restorer-postgres-13  \
golang-1-linux                        \
database-backup-restorer"

WORKDIR /release/packages
RUN bash -c 'for pkg in ${PACKAGES}; do                   \
 export BOSH_INSTALL_TARGET="/var/vcap/packages/${pkg}" ; \
 cd "/release/packages/${pkg}"                          ; \
 mkdir -p "${BOSH_INSTALL_TARGET}"                      ; \
 chmod +x packaging                                     ; \
 ./packaging                                            ; \
done'


WORKDIR /release/jobs
RUN mkdir -p /var/vcap/jobs/database-backup-restorer/bin \
 && cp database-backup-restorer/templates/backup /var/vcap/jobs/database-backup-restorer/bin/backup \
 && cp database-backup-restorer/templates/restore /var/vcap/jobs/database-backup-restorer/bin/restore \
 && chmod +x /var/vcap/jobs/database-backup-restorer/bin/backup \
 && chmod +x /var/vcap/jobs/database-backup-restorer/bin/restore

#########################################################################
# Ready to run, clean, filesystem with the stemcell and compiled packages in place
#########################################################################
FROM scratch
COPY --from=uncompressed-stemcell /stemcell/ /
COPY --from=compiled-release /var/vcap/ /var/vcap/
ENTRYPOINT ["echo", "Successfully dockerized release!"]

