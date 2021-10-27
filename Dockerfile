FROM openjdk:11-jre-slim as builder

ENV KEYCLOAK_VERSION 15.0.2

RUN apt-get update && apt-get install wget -y

RUN wget -q https://github.com/keycloak/keycloak/releases/download/${KEYCLOAK_VERSION}/keycloak-${KEYCLOAK_VERSION}.tar.gz

RUN mkdir keycloak && tar xf keycloak-${KEYCLOAK_VERSION}.tar.gz -C /keycloak --strip-components=1


WORKDIR /keycloak

#create admin user
RUN ./bin/add-user-keycloak.sh -r master -u admin -p admin --roles admin

EXPOSE 8080
EXPOSE 8787

ENTRYPOINT ["./bin/standalone.sh", "-b", "0.0.0.0", "--debug", "*:8787", "-Dkeycloak.migration.action=import", "-Dkeycloak.migration.provider=dir", "-Dkeycloak.migration.dir=/import"]