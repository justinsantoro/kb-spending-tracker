FROM golang:stretch As builder

RUN go get github.com/justinsantoro/kb-spending-tracker
RUN go build github.com/justinsantoro/kb-spending-tracker /kst

FROM keybaseio/client

ENV KEYBASE_SERVICE=1
ENV KFT_KBHOME /home/keybase
ENV KFT_KBLOC /usr/bin/keybase

COPY --from=builder /kft /home/keybase/kst

#run the bot
CMD ["/home/keybase/kst"]

#$ docker run --rm \
 #    -e KEYBASE_USERNAME="botname" \
 #    -e KEYBASE_PAPERKEY="paper key" \
 #    -e KFT_USERS="username1 username2" \
 #    -e KFT_DBGCONV="1234567" \
 #    yournewimage