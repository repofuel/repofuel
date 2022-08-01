FROM node:12.18 AS builder

WORKDIR /root

COPY . .

RUN cd web \
    && npm install \
    && npm run build \
    && rm -rf ./node_modules


FROM nginx:1.17.5-alpine

COPY web/nginx /etc/nginx
# Copy the config file to be used as a template by envsubst when running the container.
# This is a workaround to use environment variables in the Nginx configuration.
RUN cp /etc/nginx/conf.d/default.conf /etc/nginx/conf.d/default.template

COPY --from=builder /root/web/build /usr/share/nginx/html

ENV NGINX_INGEST_SRV http://ingest:3002
ENV NGINX_ACCOUNTS_SRV http://accounts:3003
ENV NGINX_AI_SRV http://ai:3004

ENTRYPOINT envsubst '\
    ${NGINX_ACCOUNTS_SRV} \
    ${NGINX_INGEST_SRV} \
    ${NGINX_AI_SRV}' \
    < /etc/nginx/conf.d/default.template > /etc/nginx/conf.d/default.conf \
    && exec nginx -g 'daemon off;'
