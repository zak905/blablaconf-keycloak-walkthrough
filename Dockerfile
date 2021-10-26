FROM alpine:3.6 as builder

ENV KEYCLOAK_VERSION 15.0.2

RUN apk add --no-cache ca-certificates wget tar

RUN wget -q https://github.com/keycloak/keycloak/releases/download/${KEYCLOAK_VERSION}/keycloak-${KEYCLOAK_VERSION}.tar.gz

RUN mkdir keycloak && tar xf keycloak-${KEYCLOAK_VERSION}.tar.gz -C keycloak --strip-components=1

FROM openjdk:11-jre-slim

COPY --from=builder /keycloak keycloak

WORKDIR keycloak

#create admin user
RUN ./bin/add-user-keycloak.sh -r master -u admin -p admin --roles admin

EXPOSE 8080

ENTRYPOINT ["./bin/standalone.sh", "-b", "0.0.0.0"]