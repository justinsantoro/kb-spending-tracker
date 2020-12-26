FROM keybaseio/client:stable-slim

ENV KEYBASE_SERVICE=1
ENV KST_KBHOME /home/keybase
ENV KST_KBLOC /usr/bin/keybase

COPY kb-spending-tracker /home/keybase/kb-spending-tracker

#run the bot
CMD ["/home/keybase/kb-spending-tracker"]

 #$ docker run --rm \
 #     -e KEYBASE_USERNAME="botname" \
 #     -e KEYBASE_PAPERKEY="paper key" \
 #     -e KST_USERS="username1,username2" \
 #     -e KST_DBGCONV="1234567" \
 #     -e KST_DBLOC="/Location/Of/database.db" \
 #     -e TZ=America/New_York \
 #     justinsantoro/kst:latest
