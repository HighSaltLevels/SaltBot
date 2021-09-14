FROM ubuntu:20.04

COPY requirements.txt saltbot /saltbot/

RUN cd /saltbot && \
    apt-get update && \ 
	apt-get install --no-install-recommends -y \
        build-essential \
        python3-dev \
        python3-pip \
        python3-setuptools \
        python3-wheel && \
    pip3 install -r requirements.txt

WORKDIR /saltbot

CMD python3 /saltbot
