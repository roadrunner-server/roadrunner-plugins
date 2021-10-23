FROM ghcr.io/spiral/roadrunner:2.5.1@sha256:e9f87866845479ede406d441a5f949adf0d05ef67a29615851021561326224c3 AS roadrunner
FROM php:8.0-cli
RUN apt update -y
RUN apt install build-essential -y


COPY --from=roadrunner /usr/bin/rr /usr/local/bin/rr
