FROM node:lts-alpine3.13
MAINTAINER Usacloud Members<sacloud.users@gmail.com>

ENV LC_ALL C.UTF-8
ENV LANG C.UTF-8

RUN npm install -g --no-cache textlint \
    textlint-filter-rule-whitelist \
    textlint-filter-rule-comments \
    textlint-rule-common-misspellings \
    textlint-rule-preset-ja-technical-writing \
    textlint-rule-preset-jtf-style

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]