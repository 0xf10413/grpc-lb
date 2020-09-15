FROM openjdk:11

RUN mkdir /app
WORKDIR /app

COPY target/grpc-lb-0.0.1-jar-with-dependencies.jar /app
ENV MAX_TRANSACTIONS=20
CMD java -jar grpc-lb-0.0.1-jar-with-dependencies.jar
