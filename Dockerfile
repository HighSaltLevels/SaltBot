FROM ubuntu:18.04

COPY requirements.txt saltbot /saltbot/

RUN cd /saltbot && \
    apt-get update && \ 
	apt-get install --no-install-recommends -y \
        build-essential \
        python3-pip \
        python3.8-dev && \
    python3.8 -m pip install setuptools==39.0.1 && \
    python3.8 -m pip install -r requirements.txt

WORKDIR /saltbot

CMD python3.8 /saltbot
