FROM haskell:9.4.8 AS freestbuilder
WORKDIR /freestBIN 
RUN git clone --branch fopss26 --depth 1 https://github.com/freest-lang/freest.git
RUN cd freest && stack install

FROM golang:1.27 AS gobuilder
WORKDIR /app
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/main.go .
RUN go build -o runFreeST .

FROM golang:1.27
WORKDIR /app
ENV LANG=C.UTF-8
COPY --from=freestbuilder /root/.local/bin/freest /usr/local/bin/freest
# Hash must be altered if the stack.yaml or GHC version is changed
COPY --from=freestbuilder /freestBIN/freest/.stack-work/install/x86_64-linux/26db9ca6e153da2ac41d8ee4b31b44e1fd0966e530403d5c06178b45d70f31d7/9.4.8/share/x86_64-linux-ghc-9.4.8/freest-5.0 /usr/local/share/freest-5.0
ENV freest_datadir=/usr/local/share/freest-5.0
COPY --from=gobuilder /app/runFreeST /app/runFreeST
#Create a new user
RUN groupadd -r freestgroup && useradd -r -m -g freestgroup freest
#Give only the necessary permissions to this user
RUN chown -R freest:freestgroup /app /usr/local/share/freest-5.0 /usr/local/bin/freest
USER freest
CMD ["/app/runFreeST"]