FROM alpine:latest

RUN apk update
RUN apk upgrade

RUN apk add nginx

COPY ./nginx.conf /etc/nginx/nginx.conf

EXPOSE 80/tcp

CMD ["nginx"]