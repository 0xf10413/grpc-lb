FROM openjdk:11

RUN mkdir /app
WORKDIR /app

COPY target/grpclb-0.0.1-jar-with-dependencies.jar /app
ENV MAX_CONNECTION_DURATION=120
CMD java -jar grpclb-0.0.1-jar-with-dependencies.jar
