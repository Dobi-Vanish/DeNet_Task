FROM alpine:latest

RUN mkdir /app


COPY rewardApp /app
COPY cmd/api/migrations /app/migrations
COPY example.env /app/example.env


WORKDIR /app

CMD [ "/app/rewardApp"]